package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/terraform"
	tf "github.com/snyk/driftctl/enumeration/terraform"
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

	SkipCredsValidation     bool `cty:"skip_credentials_validation"`
	SkipGetEC2Platforms     bool
	SkipRegionValidation    bool
	SkipRequestingAccountId bool `cty:"skip_requesting_account_id"`
	SkipMetadataApiCheck    bool
	S3ForcePathStyle        bool
}

type AWSTerraformProvider struct {
	*terraform.TerraformProvider
	session   *session.Session
	name      string
	version   string
	accountId string
}

func NewAWSTerraformProvider(version string, progress enumeration.ProgressCounter, configDir string) (*AWSTerraformProvider, error) {
	if version == "" {
		version = "3.19.0"
	}
	p := &AWSTerraformProvider{
		version: version,
		name:    "aws",
	}
	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:       p.name,
		Version:   version,
		ConfigDir: configDir,
	})
	if err != nil {
		return nil, err
	}

	p.session, err = session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}

	tfProvider, err := terraform.NewTerraformProvider(installer, terraform.TerraformProviderConfig{
		Name:         p.name,
		DefaultAlias: *p.session.Config.Region,
		GetProviderConfig: func(alias string) interface{} {
			return awsConfig{
				Region: alias,
				// Those two parameters are used to make sure that the credentials are not validated when calling
				// Configure(). Credentials validation is now handled directly in driftctl
				SkipCredsValidation:     true,
				SkipRequestingAccountId: true,

				MaxRetries: 10, // TODO make this configurable
			}
		},
	}, progress)
	if err != nil {
		return nil, err
	}
	p.TerraformProvider = tfProvider
	return p, err
}

func (a *AWSTerraformProvider) Name() string {
	return a.name
}

func (p *AWSTerraformProvider) Version() string {
	return p.version
}

var AWSCredentialsNotFoundError = errors.New("Could not find a way to authenticate on AWS!\n" +
	"Please refer to AWS documentation: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html")

func (p *AWSTerraformProvider) CheckCredentialsExist() error {
	_, err := p.session.Config.Credentials.Get()
	if err == credentials.ErrNoValidProvidersFoundInChain {
		return AWSCredentialsNotFoundError
	}
	if err != nil {
		return err
	}
	// This call is to make sure that the credentials are valid
	// A more complex logic exist in terraform provider, but it's probably not worth to implement it
	// https://github.com/hashicorp/terraform-provider-aws/blob/e3959651092864925045a6044961a73137095798/aws/auth_helpers.go#L111
	identity, err := sts.New(p.session).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		logrus.Debug(err)
		return errors.New("Could not authenticate successfully on AWS with the provided credentials.\n" +
			"Please refer to the AWS documentation: https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html\n")
	}

	p.accountId = aws.StringValue(identity.Account)
	return nil
}
