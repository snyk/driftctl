package remote

import (
	"errors"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/terraform"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoadBalancer(t *testing.T) {
	tests := []struct {
		test           string
		mocks          func(*repository.MockELBV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no load balancer",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elbv2.LoadBalancer{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "should list load balancers",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elbv2.LoadBalancer{
					{
						LoadBalancerArn:  awssdk.String("arn:aws:elasticloadbalancing:us-east-1:533948124879:loadbalancer/app/acc-test-lb-tf/9114c60e08560420"),
						LoadBalancerName: awssdk.String("acc-test-lb-tf"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "arn:aws:elasticloadbalancing:us-east-1:533948124879:loadbalancer/app/acc-test-lb-tf/9114c60e08560420", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLoadBalancerResourceType, got[0].ResourceType())
			},
		},
		{
			test: "cannot list load balancers",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllLoadBalancers").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLoadBalancerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLoadBalancerResourceType, resourceaws.AwsLoadBalancerResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
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
			fakeRepo := &repository.MockELBV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ELBV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewLoadBalancerEnumerator(repo, factory))

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
