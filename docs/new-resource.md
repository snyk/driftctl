# Add new resources

First you need to understand how driftctl scan works. Here you'll find a global overview of the step that compose the scan:

![Diagram](media/generalflow.png)

And here you'll see a more detailed flow of the retrieving resource sequence from remote:
![Diagram](media/resource.png)

## Defining the resource

First step would be to add a file under `pkg/resource/<providername>/resourcetype.go`.
This file will define a const string that will be the resource type identifier in driftctl.
Optionally, if your resource is to be supported by driftctl experimental deep mode, you can add a function that will be
applied to this resource when it's created. This allows to prevent useless diff to be displayed.
You can also add some metadata to fields so they are compared or displayed differently.

For example this defines the `aws_iam_role` resource :
```go
const AwsIamRoleResourceType = "aws_iam_role"

func initAwsIAMRoleMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	// assume_role_policy drifts will be displayed as json
	resourceSchemaRepository.UpdateSchema(AwsIamRoleResourceType, map[string]func(attributeSchema *resource.AttributeSchema){
		"assume_role_policy": func(attributeSchema *resource.AttributeSchema) {
			attributeSchema.JsonString = true
		},
	})
	// force_detach_policies should not be compared so it will be removed before the comparison
	resourceSchemaRepository.SetNormalizeFunc(AwsIamRoleResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"force_detach_policies"})
	})
}
```

When it's done you'll have to add this function to the metadata initialisation located in `pkg/resource/<providername>/metadatas.go` :
```go
func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
    initAwsAmiMetaData(resourceSchemaRepository)
}
```

In order for you new resource to be supported by our terraform state reader you should add it in `pkg/resource/resource_types.go` inside the `supportedTypes` slice.
```go
var supportedTypes = map[string]struct{}{
    "aws_ami":                               {},
}
```


All resources inside driftctl are `resource.Resource` structs.
All the other attributes are represented inside a `map[string]interface`

## Repository, Enumerator and DetailsFetcher

Then you will have to implement two interfaces:

- Repositories are the way we decided to hide direct calls to sdk and pagination logic. It's a common abstraction pattern for data retrieval.
- `remote.comon.Enumerator` is used to read resources list. It will call the cloud provider SDK to get the list of resources.
  For some resource it could make other call to enrich the resource with additional attributes when driftctl is used in deep mode
- `remote.comon.DetailsFetcher` is used to retrieve resource details. It makes a call to terraform provider `ReadResource`.
  This implementation is optional and is only needed if your resource type is to be supported by experimental deep mode.
  Please also note that it exists a generic implementation as `remote.common.GenericDetailsFetcher` that can be used with most resource type.


### Repository

This will be the component that hide all the logic linked to your provider sdk. All provider have different way to implement pagination or to name function in their api.
Here we will name all listing function `ListAll<ResourceTypeName>`.

For aws we decided to split repositories using the amazon logic. So you'll find repositories for EC2, S3 and so on.
Some provider does not have this grouping logic. Keep in mind that like all our file/struct repositories should not be too big.
So it might be useful to create a grouping logic.

For our Github implementation the number of listing function was not that heavy so we created a unique repository for everything:

```go
type GithubRepository interface {
	ListRepositories() ([]string, error)
	ListTeams() ([]Team, error)
	ListMembership() ([]string, error)
	ListTeamMemberships() ([]string, error)
	ListBranchProtection() ([]string, error)
}

type githubRepository struct {
	client GithubGraphQLClient
	ctx    context.Context
	config githubConfig
	cache  cache.Cache
}

func NewGithubRepository(config githubConfig, c cache.Cache) *githubRepository {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	oauthClient := oauth2.NewClient(ctx, ts)

	repo := &githubRepository{
		client: githubv4.NewClient(oauthClient),
		ctx:    context.Background(),
		config: config,
		cache:  c,
	}

	return repo
}
```

So as you can see this contains the logic to create the github client (it might be created outside the repository if it
makes sense to share it between multiple repositories). It also get a cache so every request is cached.
Driftctl sometimes needs to retrieve list of resources more than once, so we cache every request to avoid unnecessary call.

### Enumerator

This is used to build a resources list. Enumerators can be found in `pkg/remote/<providername>/<type>_enumerator.go`. It will call the cloud provider SDK to get the list of resources.

Note that at this point resources should not be entirely fetched.
Most of the resource returned by enumerator have empty attributes: they only represent type and terraform id.

**There are exception to this**:
- Sometime, you will need some more information about resources to retrieve them using the provider they should be added to the resource attribute maps.
- For some more complex cases, middleware needs more information that the id and type and in order to make classic run of driftctl coherent with a run with deep mode activated,
these informations should be fetched manually by the enumerator using the remote sdk.

Note that we use the classic repository to hide calls to the provider sdk.
You will probably need to at least add a listing function to list you new resource.

You should use an already implemented Enumerator as example.

For example when implementing ec2_instance resource you will need to add a ListAllInstances() function to `repository.EC2Repository`.
It will be called by the enumerator to retrieve the instances list.

Enumerator constructor could use these arguments:
- an instance of `Repository` that you will use to retrieved information about the resource
- the global resource factory that should always be used to create a new `resource.Resource`

Enumerator then need to implement:
- `SupportedType() resource.ResourceType` that will return the constant you defined in the type file at first step
- `Enumerate() ([]*resource.Resource, error)` that will return the resource listing.

```go
type EC2InstanceEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2InstanceEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2InstanceEnumerator {
	return &EC2InstanceEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2InstanceEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsInstanceResourceType
}

func (e *EC2InstanceEnumerator) Enumerate() ([]*resource.Resource, error) {
	instances, err := e.repository.ListAllInstances()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, len(instances))

	for _, instance := range instances {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*instance.InstanceId,
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
```

As you can see, listing error are treated in a particular way. Instead of failing and stopping the scan they will be handled, and an alert will be created.
So please don't forget to wrap these errors inside a `NewResourceListingError`.
For some provider error handling is not that coherent, so you might need to check in `pkg/remote/resource_enumeration_error_handler.go` and add a new case for your error.
You should test enumerator behavior when you do not have permission to enumerate resource, in the snippet above, `ListAllInstances` may return an `AccessDenied` error or so that should be handled.

Once the enumerator is written you have to add it to the remote init located in `pkg/remote/<providername>/init.go` :
```go
    remoteLibrary.AddEnumerator(NewEC2InstanceEnumerator(s3Repository, factory))
```

### DetailsFetcher

DetailsFetcher are only used by driftctl experimental deep mode.

This is the component that call terraform provider to retrieve the full attribute for each resource.
We do not want to reimplement what has already been done in every terraform provider, so you should not call the remote sdk there.

If `common.GenericDetailsFetcher` satisfy your needs you should always prefer using it instead of implementing a custom `DetailsFetcher` in a new struct.

The `DetailsFetcher` should also be added to `pkg/remote/<providername>/init.go` even if you use the generic version :
```go
    remoteLibrary.AddDetailsFetcher(aws.AwsEbsVolumeResourceType, common.NewGenericDetailsFetcher(aws.AwsEbsVolumeResourceType, provider, deserializer))
```


***Don't forget to add unit tests after adding a new resource.***

You can find example of "functional" tests in pkg/remote/<type>_scanner_test.go

You should also add acceptance tests if you think it makes sense, they are located next to the resource definition described at first step.

More information about test can be found in [testing documentation](testing.md)
