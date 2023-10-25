package remote

import (
	"errors"
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestELB_LoadBalancer(t *testing.T) {
	dummyError := errors.New("dummy error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockELBRepository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no load balancer",
			mocks: func(repository *repository.MockELBRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elb.LoadBalancerDescription{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "should list load balancers",
			mocks: func(repository *repository.MockELBRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elb.LoadBalancerDescription{
					{
						LoadBalancerName: awssdk.String("acc-test-lb-tf"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "acc-test-lb-tf", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsClassicLoadBalancerResourceType, got[0].ResourceType())
			},
		},
		{
			test: "cannot list load balancers",
			mocks: func(repository *repository.MockELBRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllLoadBalancers").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsClassicLoadBalancerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsClassicLoadBalancerResourceType, resourceaws.AwsClassicLoadBalancerResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "cannot list load balancers (dummy error)",
			mocks: func(repository *repository.MockELBRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: remoteerr.NewResourceScanningError(dummyError, resourceaws.AwsClassicLoadBalancerResourceType, ""),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockELBRepository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ELBRepository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewClassicLoadBalancerEnumerator(repo, factory))

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
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
