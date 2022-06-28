package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	aws2 "github.com/snyk/driftctl/enumeration/remote/aws"
	repository2 "github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	common2 "github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	terraform3 "github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestECRRepository(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockECRRepository, *mocks.AlerterInterface)
		err     error
	}{
		{
			test:    "no repository",
			dirName: "aws_ecr_repository_empty",
			mocks: func(client *repository2.MockECRRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple repositories",
			dirName: "aws_ecr_repository_multiple",
			mocks: func(client *repository2.MockECRRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{
					{RepositoryName: awssdk.String("test_ecr")},
					{RepositoryName: awssdk.String("bar")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list repository",
			dirName: "aws_ecr_repository_empty",
			mocks: func(client *repository2.MockECRRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllRepositories").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsEcrRepositoryResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEcrRepositoryResourceType, resourceaws.AwsEcrRepositoryResourceType), alerts.EnumerationPhase)).Return()
			},
			err: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform3.NewProviderLibrary()
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockECRRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.ECRRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository2.NewECRRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws2.NewECRRepositoryEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEcrRepositoryResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsEcrRepositoryResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEcrRepositoryResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestECRRepositoryPolicy(t *testing.T) {
	tests := []struct {
		test           string
		mocks          func(*repository2.MockECRRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		err            error
	}{
		{
			test: "single repository policy",
			mocks: func(client *repository2.MockECRRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{
					{RepositoryName: awssdk.String("test_ecr_repo_policy")},
					{RepositoryName: awssdk.String("test_ecr_repo_without_policy")},
				}, nil)
				client.On("GetRepositoryPolicy", &ecr.Repository{
					RepositoryName: awssdk.String("test_ecr_repo_policy"),
				}).Return(&ecr.GetRepositoryPolicyOutput{
					RegistryId:     awssdk.String("1"),
					RepositoryName: awssdk.String("test_ecr_repo_policy"),
				}, nil)
				client.On("GetRepositoryPolicy", &ecr.Repository{
					RepositoryName: awssdk.String("test_ecr_repo_without_policy"),
				}).Return(nil, &ecr.RepositoryPolicyNotFoundException{})
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, got[0].ResourceId(), "test_ecr_repo_policy")
			},
			err: nil,
		},
	}

	providerVersion := "3.19.0"
	schemaRepository := testresource.InitFakeSchemaRepository("aws", providerVersion)
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform3.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			remoteLibrary := common2.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository2.MockECRRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.ECRRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws2.NewECRRepositoryPolicyEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}
