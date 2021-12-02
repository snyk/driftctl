package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerr "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAutoscaling_LaunchConfiguration(t *testing.T) {
	tests := []struct {
		test           string
		mocks          func(*repository.MockAutoScalingRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no launch configuration",
			mocks: func(repository *repository.MockAutoScalingRepository, alerter *mocks.AlerterInterface) {
				repository.On("DescribeLaunchConfigurations").Return([]*autoscaling.LaunchConfiguration{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "multiple launch configurations",
			mocks: func(repository *repository.MockAutoScalingRepository, alerter *mocks.AlerterInterface) {
				repository.On("DescribeLaunchConfigurations").Return([]*autoscaling.LaunchConfiguration{
					{LaunchConfigurationName: awssdk.String("web_config_1")},
					{LaunchConfigurationName: awssdk.String("web_config_2")},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 2)

				assert.Equal(t, "web_config_1", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLaunchConfigurationResourceType, got[0].ResourceType())

				assert.Equal(t, "web_config_2", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsLaunchConfigurationResourceType, got[1].ResourceType())
			},
		},
		{
			test: "cannot list launch configurations",
			mocks: func(repository *repository.MockAutoScalingRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("DescribeLaunchConfigurations").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLaunchConfigurationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLaunchConfigurationResourceType, resourceaws.AwsLaunchConfigurationResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {

			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockAutoScalingRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.AutoScalingRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewLaunchConfigurationEnumerator(repo, factory))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
