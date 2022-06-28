package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	repository2 "github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	common2 "github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	terraform3 "github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudfront"
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

func TestCloudfrontDistribution(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository2.MockCloudfrontRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no cloudfront distributions",
			dirName: "aws_cloudfront_distribution_empty",
			mocks: func(repository *repository2.MockCloudfrontRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDistributions").Return([]*cloudfront.DistributionSummary{}, nil)
			},
		},
		{
			test:    "single cloudfront distribution",
			dirName: "aws_cloudfront_distribution_single",
			mocks: func(repository *repository2.MockCloudfrontRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllDistributions").Return([]*cloudfront.DistributionSummary{
					{Id: awssdk.String("E1M9CNS0XSHI19")},
				}, nil)
			},
		},
		{
			test:    "cannot list cloudfront distributions",
			dirName: "aws_cloudfront_distribution_list",
			mocks: func(repository *repository2.MockCloudfrontRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 400, "")
				repository.On("ListAllDistributions").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsCloudfrontDistributionResourceType, alerts.NewRemoteAccessDeniedAlert(common2.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsCloudfrontDistributionResourceType, resourceaws.AwsCloudfrontDistributionResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
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
			fakeRepo := &repository2.MockCloudfrontRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository2.CloudfrontRepository = fakeRepo
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
				repo = repository2.NewCloudfrontRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewCloudfrontDistributionEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsCloudfrontDistributionResourceType, common2.NewGenericDetailsFetcher(resourceaws.AwsCloudfrontDistributionResourceType, provider, deserializer))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsCloudfrontDistributionResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
