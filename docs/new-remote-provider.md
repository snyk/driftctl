# Add a new remote provider

A remote provider in Driftctl is a cloud provider like AWS, Github, GCP or Azure.
Current architecture allows to add a new provider in a few step.

## Declaring the new remote provider
First you need to create a new directory in `pkg/remote/<provider name>`. It will sit next to already implemented one like `pkg/remote/aws`.

Inside this directory you will create a `init.go`. First thing to do will be to define the remote name constant:
```go
const RemoteAWSTerraform = "aws+tf"
```

You will then create a function to init the provider and all the future resource enumerator. Best way to do would be to copy the function signature from an other provider:
```go
func Init(
	// Version required by the user
	version string,
	// Util to send alert
	alerter *alerter.Alerter,
	// Library that contains all providers
	providerLibrary *terraform.ProviderLibrary,
	// Library that contains the enumerators and details fetcher for each supported resources
	remoteLibrary *common.RemoteLibrary,
	// progress display
	progress output.Progress,
	// Repository for all resource schema
	resourceSchemaRepository *resource.SchemaRepository,
	// Factory used to create driftctl resource
	factory resource.ResourceFactory,
	// Drifctl config directory (in which terraform provider is downloaded)
	configDir string) error {

	// Define the default version of terraform provider to be used. When the user does not require a specific one
	if version == "" {
		version = "3.19.0"
	}

	// This is this actual terraform provider creation
	provider, err := NewAWSTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	// And then initialisation
	err = provider.Init()
	if err != nil {
		return err
	}

	// You'll need to create a new cache that will be use to cache fetched resources lists
	repositoryCache := cache.New(100)

	// Deserializer is used to convert cty value return by terraform provider to driftctl AbstactResource
    deserializer := resource.NewDeserializer(factory)

    // Adding the provider to the library
    providerLibrary.AddProvider(terraform.AWS, provider)
}
```

When it's done you'll create a `provider.go` file to contains your terraform provider representation. Again you should looks at other implementation :
```go
// Define your actual provider representation, It is required to compose with terraform.TerraformProvider and to have a name and a version
// Please note that the name should match the real terraform provider name.
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
	// Use terraformproviderinstaller to retreive the provider if needed
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

	// Config is dependant on the teraform provider needs.
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

You are now almost done. You'll need to make driftctl aware of this provider so in `pkg/remote/remote.go` add your new constant in `supportedRemotes`:
```go
var supportedRemotes = []string{
	aws.RemoteAWSTerraform,
	github.RemoteGithubTerraform,
}
```
And don't forget to modify the Activate function to be able to activate your new provider. You'll need to add a new case in the switch:
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

Your provider is now set up !

## Prepare Driftctl to support new resources

New resource for the just added provider will be located in `pkg/resource/<provider name>`. You should create this directory and the `metadata.go` file.
Inside this file add a new function:
```go
func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
}
```

And add a call to it in the `remote/<provider>/init.go` you created at first step.

Last step will add to create test for the new resource you will implement.
Please use TestCreateNewSchema located in `test/schemas/schemas_test.go` to generate a schema file that will be used for the mocked provider.

Everything is not ready, you should [start adding new resources](new-resource.md) !
