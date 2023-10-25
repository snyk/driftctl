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
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestELBV2_LoadBalancer(t *testing.T) {
	dummyError := errors.New("dummy error")

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
			test: "cannot list load balancers (403)",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllLoadBalancers").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLoadBalancerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLoadBalancerResourceType, resourceaws.AwsLoadBalancerResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "cannot list load balancers (dummy error)",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: remoteerr.NewResourceScanningError(dummyError, resourceaws.AwsLoadBalancerResourceType, ""),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockELBV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ELBV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewLoadBalancerEnumerator(repo, factory))

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

func TestELBV2_LoadBalancerListener(t *testing.T) {
	dummyError := errors.New("dummy error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockELBV2Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "no load balancer listener",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elbv2.LoadBalancer{
					{
						LoadBalancerArn: awssdk.String("test-lb"),
					},
				}, nil)
				repository.On("ListAllLoadBalancerListeners", "test-lb").Return([]*elbv2.Listener{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "should list load balancer listener",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elbv2.LoadBalancer{
					{
						LoadBalancerArn: awssdk.String("test-lb"),
					},
				}, nil)

				repository.On("ListAllLoadBalancerListeners", "test-lb").Return([]*elbv2.Listener{
					{
						ListenerArn: awssdk.String("test-lb-listener-1"),
					},
				}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, "test-lb-listener-1", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsLoadBalancerListenerResourceType, got[0].ResourceType())
			},
		},
		{
			test: "cannot list load balancer listeners (403)",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elbv2.LoadBalancer{
					{
						LoadBalancerArn: awssdk.String("test-lb"),
					},
				}, nil)

				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllLoadBalancerListeners", "test-lb").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLoadBalancerListenerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingError(awsError, resourceaws.AwsLoadBalancerListenerResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "cannot list load balancers (403)",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllLoadBalancers").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsLoadBalancerListenerResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsLoadBalancerListenerResourceType, resourceaws.AwsLoadBalancerResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
		{
			test: "cannot list load balancer listeners (dummy error)",
			mocks: func(repository *repository.MockELBV2Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllLoadBalancers").Return([]*elbv2.LoadBalancer{
					{
						LoadBalancerArn: awssdk.String("test-lb"),
					},
				}, nil)

				repository.On("ListAllLoadBalancerListeners", "test-lb").Return(nil, dummyError)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: remoteerr.NewResourceScanningError(dummyError, resourceaws.AwsLoadBalancerListenerResourceType, ""),
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockELBV2Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.ELBV2Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewLoadBalancerListenerEnumerator(repo, factory))

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
