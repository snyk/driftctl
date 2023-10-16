package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestECRRepository(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockECRRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		err            error
	}{
		{
			test:    "no repository",
			dirName: "aws_ecr_repository_empty",
			mocks: func(client *repository.MockECRRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
		{
			test:    "multiple repositories",
			dirName: "aws_ecr_repository_multiple",
			mocks: func(client *repository.MockECRRepository, alerter *mocks.AlerterInterface) {
				client.On("ListAllRepositories").Return([]*ecr.Repository{
					{RepositoryName: awssdk.String("test_ecr")},
					{RepositoryName: awssdk.String("bar")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "test_ecr", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsEcrRepositoryResourceType, got[0].ResourceType())

				assert.Equal(t, "bar", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsEcrRepositoryResourceType, got[1].ResourceType())
			},
			err: nil,
		},
		{
			test:    "cannot list repository",
			dirName: "aws_ecr_repository_empty",
			mocks: func(client *repository.MockECRRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllRepositories").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsEcrRepositoryResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsEcrRepositoryResourceType, resourceaws.AwsEcrRepositoryResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			err: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockECRRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ECRRepository = fakeRepo
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
				repo = repository.NewECRRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewECRRepositoryEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestECRRepositoryPolicy(t *testing.T) {
	tests := []struct {
		test           string
		mocks          func(*repository.MockECRRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		err            error
	}{
		{
			test: "single repository policy",
			mocks: func(client *repository.MockECRRepository, alerter *mocks.AlerterInterface) {
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

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockECRRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ECRRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewECRRepositoryPolicyEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
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
