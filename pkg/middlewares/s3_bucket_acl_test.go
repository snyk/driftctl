package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"

	awssdk "github.com/aws/aws-sdk-go/aws"
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
					&aws.AwsS3Bucket{
						Id: "testgrant",
						Grant: &[]struct {
							Id          string   `cty:"id"`
							Permissions []string `cty:"permissions"`
							Type        *string  `cty:"type"`
							Uri         *string  `cty:"uri"`
						}{
							{
								"356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
								[]string{"FULL_CONTROL"},
								awssdk.String("CanonicalUser"),
								awssdk.String(""),
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsAmi{},
					&aws.AwsS3Bucket{
						Id:  "testgrant",
						Acl: awssdk.String("public-read"),
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				s3Bucket, _ := (*remoteResources)[0].(*aws.AwsS3Bucket)
				assert.Empty(s3Bucket.Grant)
			},
		},
		{
			name: "does not modify grant field on remote resource if acl field is private",
			args: args{
				remoteResources: &[]resource.Resource{
					&aws.AwsS3Bucket{
						Id: "testgrant",
						Grant: &[]struct {
							Id          string   `cty:"id"`
							Permissions []string `cty:"permissions"`
							Type        *string  `cty:"type"`
							Uri         *string  `cty:"uri"`
						}{
							{
								"356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
								[]string{"FULL_CONTROL"},
								awssdk.String("CanonicalUser"),
								awssdk.String(""),
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsAmi{},
					&aws.AwsS3Bucket{
						Id:  "testgrant",
						Acl: awssdk.String("private"),
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				s3Bucket, _ := (*remoteResources)[0].(*aws.AwsS3Bucket)
				assert.Len(*s3Bucket.Grant, 1)
				expected := struct {
					Id          string   `cty:"id"`
					Permissions []string `cty:"permissions"`
					Type        *string  `cty:"type"`
					Uri         *string  `cty:"uri"`
				}{
					Id:          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
					Permissions: []string{"FULL_CONTROL"},
					Type:        awssdk.String("CanonicalUser"),
					Uri:         awssdk.String(""),
				}
				assert.Equal(expected, (*s3Bucket.Grant)[0])
			},
		},
		{
			name: "does not modify grant field on remote resource if acl field is undefined",
			args: args{
				remoteResources: &[]resource.Resource{
					&aws.AwsS3Bucket{
						Id: "testgrant",
						Grant: &[]struct {
							Id          string   `cty:"id"`
							Permissions []string `cty:"permissions"`
							Type        *string  `cty:"type"`
							Uri         *string  `cty:"uri"`
						}{
							{
								"356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
								[]string{"FULL_CONTROL"},
								awssdk.String("CanonicalUser"),
								awssdk.String(""),
							},
						},
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsAmi{},
					&aws.AwsS3Bucket{
						Id: "testgrant",
					},
				},
			},
			assert: func(assert *assert.Assertions, remoteResources, resourcesFromState *[]resource.Resource) {
				s3Bucket, _ := (*remoteResources)[0].(*aws.AwsS3Bucket)
				assert.Len(*s3Bucket.Grant, 1)
				expected := struct {
					Id          string   `cty:"id"`
					Permissions []string `cty:"permissions"`
					Type        *string  `cty:"type"`
					Uri         *string  `cty:"uri"`
				}{
					Id:          "356616ba70ebbea29732c95eef24f9ea326b9018c167651705348b5af406a6db",
					Permissions: []string{"FULL_CONTROL"},
					Type:        awssdk.String("CanonicalUser"),
					Uri:         awssdk.String(""),
				}
				assert.Equal(expected, (*s3Bucket.Grant)[0])
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
