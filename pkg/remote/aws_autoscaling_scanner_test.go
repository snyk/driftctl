package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraformtest "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAutoscaling_LaunchConfiguration(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockAutoScalingRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no launch configuration",
			dirName: "aws_launch_configuration",
			mocks: func(repository *repository.MockAutoScalingRepository, alerter *mocks.AlerterInterface) {
				repository.On("DescribeLaunchConfigurations").Return([]*autoscaling.LaunchConfiguration{}, nil)
			},
		},
		{
			test:    "multiple launch configurations",
			dirName: "aws_launch_configuration_multiple",
			mocks: func(repository *repository.MockAutoScalingRepository, alerter *mocks.AlerterInterface) {
				repository.On("DescribeLaunchConfigurations").Return([]*autoscaling.LaunchConfiguration{
					{LaunchConfigurationName: awssdk.String("web_config_1")},
					{LaunchConfigurationName: awssdk.String("web_config_2")},
				}, nil)
			},
		},
		{
			test:    "cannot list launch configurations",
			dirName: "aws_launch_configuration",
			mocks: func(repository *repository.MockAutoScalingRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("DescribeLaunchConfigurations").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLaunchConfigurationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLaunchConfigurationResourceType, resourceaws.AwsLaunchConfigurationResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockAutoScalingRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.AutoScalingRepository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraformtest.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraformtest.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewAutoScalingRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewLaunchConfigurationEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsLaunchConfigurationResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsLaunchConfigurationResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsLaunchConfigurationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}
