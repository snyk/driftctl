package middlewares

import (
	"testing"

	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/pkg/terraform"
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
			name: "should not map any rds cluster instance into db instances",
			remoteResources: []*resource.Resource{
				{
					Id:    "db-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "db-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			stateResources: []*resource.Resource{},
			expectedRemoteResources: []*resource.Resource{
				{
					Id:    "db-0",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "db-1",
					Type:  aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			expectedStateResources: []*resource.Resource{},
		},
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
					Id:   "aurora-cluster-demo-0",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
				{
					Id:   "aurora-cluster-demo-1",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
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
					Id:   "aurora-cluster-demo-0",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
				{
					Id:   "aurora-cluster-demo-1",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
			},
			expectedStateResources: []*resource.Resource{
				{
					Id:   "aurora-cluster-demo-0",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
				{
					Id:   "aurora-cluster-demo-1",
					Type: aws.AwsDbInstanceResourceType,
					Attrs: &resource.Attributes{
						"field": "test",
					},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsDbInstanceResourceType, "aurora-cluster-demo-0", map[string]interface{}{"field": "test"}).
					Return(&resource.Resource{
						Id:    "aurora-cluster-demo-0",
						Type:  aws.AwsDbInstanceResourceType,
						Attrs: &resource.Attributes{"field": "test"},
					}).
					Once()

				factory.On("CreateAbstractResource", aws.AwsDbInstanceResourceType, "aurora-cluster-demo-1", map[string]interface{}{"field": "test"}).
					Return(&resource.Resource{
						Id:    "aurora-cluster-demo-1",
						Type:  aws.AwsDbInstanceResourceType,
						Attrs: &resource.Attributes{"field": "test"},
					}).
					Once()
			},
		},
		{
			name: "should find only one db instances in remote",
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
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", aws.AwsDbInstanceResourceType, "aurora-cluster-demo-0", map[string]interface{}{}).
					Return(&resource.Resource{
						Id:    "aurora-cluster-demo-0",
						Type:  aws.AwsDbInstanceResourceType,
						Attrs: &resource.Attributes{},
					}).
					Once()
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
