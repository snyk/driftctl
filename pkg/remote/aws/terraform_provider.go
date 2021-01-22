package aws

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/sirupsen/logrus"

	tf "github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type awsConfig struct {
	AccessKey     string
	SecretKey     string
	CredsFilename string
	Profile       string
	Token         string
	Region        string `cty:"region"`
	MaxRetries    int

	AssumeRoleARN         string
	AssumeRoleExternalID  string
	AssumeRoleSessionName string
	AssumeRolePolicy      string

	AllowedAccountIds   []string
	ForbiddenAccountIds []string

	Endpoints        map[string]string
	IgnoreTagsConfig map[string]string
	Insecure         bool

	SkipCredsValidation     bool
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool
}

type TerraformProvider struct {
	lock             sync.Mutex
	providerSupplier *tf.ProviderInstaller
	session          *session.Session
	grpcProviders    map[string]*plugin.GRPCProvider
	schemas          map[string]providers.Schema
	defaultRegion    string
	runner           *parallel.ParallelRunner
}

func NewTerraFormProvider() (*TerraformProvider, error) {
	provider, err := tf.NewProviderInstaller()
	if err != nil {
		return nil, err
	}
	p := TerraformProvider{
		providerSupplier: provider,
		runner:           parallel.NewParallelRunner(context.TODO(), 10),
		grpcProviders:    make(map[string]*plugin.GRPCProvider),
	}
	p.initSession()
	p.defaultRegion = *p.session.Config.Region
	stopCh := make(chan bool)
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-c:
			logrus.Warn("Detected interrupt during terraform provider configuration, cleanup ...")
			p.Cleanup()
			os.Exit(1)
		case <-stopCh:
			return
		}
	}()
	defer func() {
		stopCh <- true
	}()
	err = p.configure(p.defaultRegion)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Scanning AWS on region: %s\n", p.defaultRegion)
	return &p, nil
}

func (p *TerraformProvider) Schema() map[string]providers.Schema {
	return p.schemas
}

func (p *TerraformProvider) Runner() *parallel.ParallelRunner {
	return p.runner
}

func (p *TerraformProvider) initSession() {
	p.session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
}

func (p *TerraformProvider) configure(region string) error {

	providerPath, err := p.providerSupplier.GetAws()
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"region": region,
	}).Debug("Starting new provider")
	if p.grpcProviders[region] == nil {
		logrus.WithFields(logrus.Fields{
			"region": region,
		}).Debug("Starting aws provider GRPC client")
		GRPCProvider, err := tf.NewTerraformProvider(discovery.PluginMeta{
			Path: providerPath,
		})

		if err != nil {
			return err
		}
		p.grpcProviders[region] = GRPCProvider
	}

	schema := p.grpcProviders[region].GetSchema()
	if p.schemas == nil {
		p.schemas = schema.ResourceTypes
	}
	configType := schema.Provider.Block.ImpliedType()
	val, err := gocty.ToCtyValue(getConfig(region), configType)
	if err != nil {
		return err
	}
	resp := p.grpcProviders[region].Configure(providers.ConfigureRequest{
		Config: val,
	})

	if resp.Diagnostics.HasErrors() {
		return resp.Diagnostics.Err()
	}

	return nil
}

func getConfig(region string) awsConfig {
	return awsConfig{
		Region:     region,
		MaxRetries: 10, // TODO make this configurable
	}
}

func (p *TerraformProvider) ReadResource(args tf.ReadResourceArgs) (*cty.Value, error) {

	logrus.WithFields(logrus.Fields{
		"id":    args.ID,
		"type":  args.Ty,
		"attrs": args.Attributes,
	}).Debugf("Reading aws cloud resource")

	typ := string(args.Ty)
	state := &terraform.InstanceState{
		ID:         args.ID,
		Attributes: map[string]string{},
	}

	region := p.defaultRegion
	if args.Attributes["aws_region"] != "" {
		region = args.Attributes["aws_region"]
		delete(args.Attributes, "aws_region")
	}

	p.lock.Lock()
	if p.grpcProviders[region] == nil {
		err := p.configure(region)
		if err != nil {
			return nil, err
		}
	}
	p.lock.Unlock()

	if args.Attributes != nil && len(args.Attributes) > 0 {
		// call to the provider sometimes add and delete field to their attribute this may broke caller so we deep copy attributes
		state.Attributes = make(map[string]string, len(args.Attributes))
		for k, v := range args.Attributes {
			state.Attributes[k] = v
		}
	}

	impliedType := p.schemas[typ].Block.ImpliedType()

	priorState, err := state.AttrsAsObjectValue(impliedType)
	if err != nil {
		return nil, err
	}

	var newState cty.Value
	r := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil)

	err = r.Run(func() error {
		resp := p.grpcProviders[region].ReadResource(providers.ReadResourceRequest{
			TypeName:     typ,
			PriorState:   priorState,
			Private:      []byte{},
			ProviderMeta: cty.NullVal(cty.DynamicPseudoType),
		})
		if resp.Diagnostics.HasErrors() {
			return resp.Diagnostics.Err()
		}
		nonFatalErr := resp.Diagnostics.NonFatalErr()
		if resp.NewState.IsNull() && nonFatalErr != nil {
			return errors.Errorf("state returned by ReadResource is nil: %+v", nonFatalErr)
		}
		newState = resp.NewState
		return nil
	})

	if err != nil {
		return nil, err
	}
	return &newState, nil
}

func (p *TerraformProvider) Cleanup() {
	for region, client := range p.grpcProviders {
		logrus.WithFields(logrus.Fields{
			"region": region,
		}).Debug("Closing gRPC client")
		client.Close()
	}
}
