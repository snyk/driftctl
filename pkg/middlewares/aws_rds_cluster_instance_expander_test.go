package middlewares

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAwsRDSClusterInstanceExpander_Execute(t *testing.T) {
	tests := []struct {
		name                    string
		remoteResources         []*resource.Resource
		stateResources          []*resource.Resource
		expectedRemoteResources []*resource.Resource
		expectedStateResources  []*resource.Resource
		mock                    func(factory *terraform.MockResourceFactory)
	}{
		{
			name: "should import db instances in state",
			remoteResources: []*resource.Resource{
				{
					Id:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			stateResources: []*resource.Resource{
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-1",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedRemoteResources: []*resource.Resource{
				{
					Id:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedStateResources: []*resource.Resource{
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "should not find db instances in remote",
			remoteResources: []*resource.Resource{
				{
					Id:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			stateResources: []*resource.Resource{
				{
					Id:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-1",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedRemoteResources: []*resource.Resource{
				{
					Id:    "bucket89713",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedStateResources: []*resource.Resource{
				{
					Id:    "bucket01",
					Type:  aws.AwsS3BucketResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "aurora-cluster-demo-1",
					Type:  aws.AwsRDSClusterInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewRDSClusterInstanceExpander(factory)
			err := m.Execute(&tt.remoteResources, &tt.stateResources)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expectedRemoteResources, tt.remoteResources, "Unexpected remote resources")
			assert.Equal(t, tt.expectedStateResources, tt.stateResources, "Unexpected state resources")
		})
	}
}
