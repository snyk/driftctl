package remote

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCloudfrontDistribution(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockCloudfrontRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no cloudfront distributions",
			dirName: "aws_cloudfront_distribution_empty",
			mocks: func(repository *repository.MockCloudfrontRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDistributions").Return([]*cloudfront.DistributionSummary{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test:    "single cloudfront distribution",
			dirName: "aws_cloudfront_distribution_single",
			mocks: func(repository *repository.MockCloudfrontRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDistributions").Return([]*cloudfront.DistributionSummary{
					{Id: awssdk.String("E1M9CNS0XSHI19")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)

				assert.Equal(t, "E1M9CNS0XSHI19", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsCloudfrontDistributionResourceType, got[0].ResourceType())
			},
		},
		{
			test:    "cannot list cloudfront distributions",
			dirName: "aws_cloudfront_distribution_list",
			mocks: func(repository *repository.MockCloudfrontRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, "")
				repository.On("ListAllDistributions").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsCloudfrontDistributionResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsCloudfrontDistributionResourceType, resourceaws.AwsCloudfrontDistributionResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
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
			fakeRepo := &repository.MockCloudfrontRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.CloudfrontRepository = fakeRepo
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
				repo = repository.NewCloudfrontRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewCloudfrontDistributionEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
