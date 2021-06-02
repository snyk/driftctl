package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestS3BucketAcl_Execute(t *testing.T) {
	type args struct {
		remoteResources    *[]resource.Resource
		resourcesFromState *[]resource.Resource
	}
	tests := []struct {
		name   string
		args   args
		assert func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource)
	}{
		{
			name: "grant field on remote resource must be reset if acl != private in state resource",
			args: args{
				remoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"grant": []map[string]interface{}{
								{
									"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
									"permissions": []string{"FULL_CONTROL"},
									"type":        "CanonicalUser",
									"uri":         "",
								},
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Type: aws.AwsAmiResourceType,
					},
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"acl": "public-read",
						},
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				remoteRes, _ := (*remoteResources)[0].(*resource.AbstractResource)
				stateRes, _ := (*resourcesFromState)[1].(*resource.AbstractResource)
				_, exist := remoteRes.Attrs.Get("grant")
				_, stateAclExist := stateRes.Attrs.Get("acl")
				_, remoteAclExist := remoteRes.Attrs.Get("acl")
				assert.False(exist)
				assert.False(stateAclExist)
				assert.False(remoteAclExist)
			},
		},
		{
			name: "does not modify grant field on remote resource if acl field is private",
			args: args{
				remoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"grant": []map[string]interface{}{
								{
									"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
									"permissions": []string{"FULL_CONTROL"},
									"type":        "CanonicalUser",
									"uri":         "",
								},
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Type: aws.AwsAmiResourceType,
					},
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"acl": "private",
						},
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				s3Bucket, _ := (*remoteResources)[0].(*resource.AbstractResource)
				grantAttr, exist := s3Bucket.Attrs.Get("grant")
				grant := grantAttr.([]map[string]interface{})
				assert.True(exist)
				assert.Len(grant, 1)
				expected := map[string]interface{}{
					"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
					"permissions": []string{"FULL_CONTROL"},
					"type":        "CanonicalUser",
					"uri":         "",
				}
				assert.Equal(expected, grant[0])
			},
		},
		{
			name: "does not modify grant field on remote resource if acl field is undefined",
			args: args{
				remoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"grant": []map[string]interface{}{
								{
									"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
									"permissions": []string{"FULL_CONTROL"},
									"type":        "CanonicalUser",
									"uri":         "",
								},
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Type: aws.AwsAmiResourceType,
					},
					&resource.AbstractResource{
						Id:    "testgrant",
						Type:  aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{},
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				s3Bucket, _ := (*remoteResources)[0].(*resource.AbstractResource)
				grantAttr, exist := s3Bucket.Attrs.Get("grant")
				grant := grantAttr.([]map[string]interface{})
				assert.True(exist)
				assert.Len(grant, 1)
				expected := map[string]interface{}{
					"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
					"permissions": []string{"FULL_CONTROL"},
					"type":        "CanonicalUser",
					"uri":         "",
				}
				assert.Equal(expected, grant[0])
			},
		},
		{
			name: "does not modify grant field on remote resource if acl field is empty",
			args: args{
				remoteResources: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"grant": []map[string]interface{}{
								{
									"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
									"permissions": []string{"FULL_CONTROL"},
									"type":        "CanonicalUser",
									"uri":         "",
								},
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&resource.AbstractResource{
						Type: aws.AwsAmiResourceType,
					},
					&resource.AbstractResource{
						Id:   "testgrant",
						Type: aws.AwsS3BucketResourceType,
						Attrs: &resource.Attributes{
							"acl": "",
						},
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				s3Bucket, _ := (*remoteResources)[0].(*resource.AbstractResource)
				grantAttr, exist := s3Bucket.Attrs.Get("grant")
				grant := grantAttr.([]map[string]interface{})
				assert.True(exist)
				assert.Len(grant, 1)
				expected := map[string]interface{}{
					"id":          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
					"permissions": []string{"FULL_CONTROL"},
					"type":        "CanonicalUser",
					"uri":         "",
				}
				assert.Equal(expected, grant[0])
			},
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(tt *testing.T) {
			assert := assert.New(tt)
			m := S3BucketAcl{}
			if err := m.Execute(c.args.remoteResources, c.args.resourcesFromState); err != nil {
				tt.Error(err)
			}
			c.assert(assert, c.args.remoteResources, c.args.resourcesFromState)
		})
	}
}
