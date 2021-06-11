package terraform

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/eapache/go-resiliency/retrier"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
	"github.com/hashicorp/terraform/providers"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	tf "github.com/cloudskiff/driftctl/pkg/terraform"
)

// "alias" in these struct are a way to namespace gRPC clients.
// For example, if we need to read S3 bucket from multiple AWS region,
// we'll have an alias per region, and the alias IS the region itself.
// So we can query resources using a specific custom provider configuration
type TerraformProviderConfig struct {
	Name              string
	DefaultAlias      string
	GetProviderConfig func(alias string) interface{}
}

type TerraformProvider struct {
	lock              sync.Mutex
	providerInstaller *tf.ProviderInstaller
	grpcProviders     map[string]*plugin.GRPCProvider
	schemas           map[string]providers.Schema
	Config            TerraformProviderConfig
	runner            *parallel.ParallelRunner
	progress          output.Progress
	driftIgnore       *filter.DriftIgnore
}

func NewTerraformProvider(installer *tf.ProviderInstaller,
	config TerraformProviderConfig,
	progress output.Progress,
	driftIgnore *filter.DriftIgnore) (*TerraformProvider, error) {
	p := TerraformProvider{
		providerInstaller: installer,
		runner:            parallel.NewParallelRunner(context.TODO(), 10),
		grpcProviders:     make(map[string]*plugin.GRPCProvider),
		Config:            config,
		progress:          progress,
		driftIgnore:       driftIgnore,
	}
	return &p, nil
}

func (p *TerraformProvider) Init() error {
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
	err := p.configure(p.Config.DefaultAlias)
	if err != nil {
		return err
	}
	return nil
}

func (p *TerraformProvider) Schema() map[string]providers.Schema {
	return p.schemas
}

func (p *TerraformProvider) Runner() *parallel.ParallelRunner {
	return p.runner
}

func (p *TerraformProvider) configure(alias string) error {

	providerPath, err := p.providerInstaller.Install()
	if err != nil {
		return err
	}

	if p.grpcProviders[alias] == nil {
		logrus.WithFields(logrus.Fields{
			"alias": alias,
		}).Debug("Starting gRPC client")
		GRPCProvider, err := tf.NewGRPCProvider(discovery.PluginMeta{
			Path: providerPath,
		})

		if err != nil {
			return err
		}
		p.grpcProviders[alias] = GRPCProvider
	}

	schema := p.grpcProviders[alias].GetSchema()
	if p.schemas == nil {
		p.schemas = schema.ResourceTypes
	}
	configType := schema.Provider.Block.ImpliedType()
	val, err := gocty.ToCtyValue(p.Config.GetProviderConfig(alias), configType)
	if err != nil {
		return err
	}
	resp := p.grpcProviders[alias].Configure(providers.ConfigureRequest{
		Config: val,
	})

	if resp.Diagnostics.HasErrors() {
		return resp.Diagnostics.Err()
	}

	logrus.WithFields(logrus.Fields{
		"alias": alias,
	}).Debug("New gRPC client started")

	logrus.WithFields(logrus.Fields{
		"name":  p.Config.Name,
		"alias": alias,
	}).Debug("Terraform provider initialized")

	return nil
}

func (p *TerraformProvider) ReadResource(args tf.ReadResourceArgs) (*cty.Value, error) {

	logrus.WithFields(logrus.Fields{
		"id":    args.ID,
		"type":  args.Ty,
		"attrs": args.Attributes,
	}).Debugf("Reading cloud resource")
	typ := string(args.Ty)

	if p.driftIgnore.IsResourceIgnored(&resource.AbstractResource{
		Id:   args.ID,
		Type: typ,
	}) {
		logrus.WithFields(logrus.Fields{
			"id":    args.ID,
			"type":  args.Ty,
			"attrs": args.Attributes,
		}).Debugf("Resource is ignored in drifignore, skipping...")
		return &cty.NilVal, nil
	}

	state := &terraform.InstanceState{
		ID:         args.ID,
		Attributes: map[string]string{},
	}

	alias := p.Config.DefaultAlias
	if args.Attributes["alias"] != "" {
		alias = args.Attributes["alias"]
		delete(args.Attributes, "alias")
	}

	p.lock.Lock()
	if p.grpcProviders[alias] == nil {
		err := p.configure(alias)
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
		resp := p.grpcProviders[alias].ReadResource(providers.ReadResourceRequest{
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
	p.progress.Inc()
	return &newState, nil
}

func (p *TerraformProvider) Cleanup() {
	for alias, client := range p.grpcProviders {
		logrus.WithFields(logrus.Fields{
			"alias": alias,
		}).Debug("Closing gRPC client")
		client.Close()
	}
}
