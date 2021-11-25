# Add a new remote provider

A remote provider in driftctl represents a cloud provider like AWS, GitHub, GCP or Azure.

Our current architecture allows to add a new provider in a few steps.

## Declaring the new remote provider

First you need to create a new directory in `pkg/remote/<provider name>`. It will sit next to already implemented one like `pkg/remote/aws`.

Inside this directory, you will create a `init.go` file in which you will define the remote name constant:

```go
const RemoteAWSTerraform = "aws+tf"
```

`+tf` means that we use Terraform to retrieve resource's details, in the future, we may add other ways to read those details.

You will then create a function to initialize the provider and all resource's enumerators. The best way to do it would be to copy the function signature from another provider:

```go
func Init(
	// Version required by the user
	version string,
	// Util to send alert
	alerter *alerter.Alerter,
	// Library that contains all providers
	providerLibrary *terraform.ProviderLibrary,
	// Library that contains enumerators and details fetchers for each supported resources
	remoteLibrary *common.RemoteLibrary,
	// Progress displayer
	progress output.Progress,
	// Repository for all resource schemas
	resourceSchemaRepository *resource.SchemaRepository,
	// Factory used to create driftctl resource
	factory resource.ResourceFactory,
	// driftctl configuration directory (where Terraform provider is downloaded)
	configDir string) error {

	// You need to define the default version of the Terraform provider when the user does not specify one
	if version == "" {
		version = "3.19.0"
	}

	// Creation of the Terraform provider
	provider, err := NewAWSTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	// And then initialization
	err = provider.Init()
	if err != nil {
		return err
	}

	// You'll need to create a new cache that will be used to cache fetched lists of resources
	repositoryCache := cache.New(100)

	// Deserializer is used to convert cty value returned by Terraform provider to driftctl Resource
    deserializer := resource.NewDeserializer(factory)

    // Adding the provider to the library
    providerLibrary.AddProvider(terraform.AWS, provider)
}
```

Once done, you'll create a `provider.go` file to contain your Terraform provider representation. Again you should look at other implementation:

```go
// Define your actual provider representation, it is required to compose with terraform.TerraformProvider, a name and a version
// Please note that the name should match the real Terraform provider name.
type AWSTerraformProvider struct {
	*terraform.TerraformProvider
	session *session.Session
	name    string
	version string
}

func NewAWSTerraformProvider(version string, progress output.Progress, configDir string) (*AWSTerraformProvider, error) {
	// Just pass your version and name
	p := &AWSTerraformProvider{
		version: version,
		name:    "aws",
	}
	// Use Terraform ProviderInstaller to retrieve the provider if needed
	installer, err := tf.NewProviderInstaller(tf.ProviderConfig{
		Key:       p.name,
		Version:   version,
		ConfigDir: configDir,
	})
	if err != nil {
		return nil, err
	}
	p.session = session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// ProviderConfig is dependent on the Terraform provider needs.
	tfProvider, err := terraform.NewTerraformProvider(installer, terraform.TerraformProviderConfig{
		Name:         p.name,
		DefaultAlias: *p.session.Config.Region,
		GetProviderConfig: func(alias string) interface{} {
			return awsConfig{
				Region:     alias,
				MaxRetries: 10,
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
```

The configuration returned in `GetProviderConfig` should be annotated with `cty` tags to be passed to the provider.

```go
type githubConfig struct {
	Token        string
	Owner        string `cty:"owner"`
	Organization string
}
```

You are now almost done. You'll need to make driftctl aware of this provider. Thus, the in `pkg/remote/remote.go` file, add your new constant in `supportedRemotes`:

```go
var supportedRemotes = []string{
	aws.RemoteAWSTerraform,
	github.RemoteGithubTerraform,
}
```

Don't forget to modify the Activate function. You'll need to add a new case in the switch statement:

```go
func Activate(remote, version string, alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	remoteLibrary *common.RemoteLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory,
	configDir string) error {
	switch remote {
	case aws.RemoteAWSTerraform:
		return aws.Init(version, alerter, providerLibrary, remoteLibrary, progress, resourceSchemaRepository, factory, configDir)
	case github.RemoteGithubTerraform:
		return github.Init(version, alerter, providerLibrary, remoteLibrary, progress, resourceSchemaRepository, factory, configDir)
	default:
		return errors.Errorf("unsupported remote '%s'", remote)
	}
}
```

Your provider is now set up!

## Prepare driftctl to support new resources

Each new resource of the newly added provider will be located in `pkg/resource/<provider name>` directory. You need to create the latter and the `metadatas.go` file inside it.

Inside this file add a new function:

```go
func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
}
```

Then, add a call to this function in the `remote/<provider>/init.go` file you created in the first step.

You also need to create a test schema for upcoming tests.

Please use `TestCreateNewSchema` located in `test/terraform/schemas_test.go` to generate a schema file that will be used for the mocked provider.

Everything is now ready, you should [start adding new resources](new-resource.md)!
