package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/r3labs/diff/v2"
)

func TestAwsInstanceBlockDeviceResourceMapper_Execute(t *testing.T) {
	type args struct {
		expectedResource   *[]resource.Resource
		resourcesFromState *[]resource.Resource
	}
	tests := []struct {
		name    string
		args    args
		mocks   func(factory *terraform.MockResourceFactory)
		wantErr bool
	}{
		{
			"Test with root block device and ebs block device",
			struct {
				expectedResource   *[]resource.Resource
				resourcesFromState *[]resource.Resource
			}{
				expectedResource: &[]resource.Resource{
					&resource.AbstractResource{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
							"volume_tags": map[string]string{
								"Name": "rootVol",
							},
						},
					},
					&aws.AwsEbsVolume{
						AvailabilityZone:   awssdk.String("eu-west-3"),
						Encrypted:          awssdk.Bool(true),
						Id:                 "vol-02862d9b39045a3a4",
						Iops:               awssdk.Int(1234),
						KmsKeyId:           awssdk.String("kms"),
						Size:               awssdk.Int(8),
						Type:               awssdk.String("gp2"),
						MultiAttachEnabled: awssdk.Bool(false),
						Tags: map[string]string{
							"Name": "rootVol",
						},
					},
					&aws.AwsEbsVolume{
						AvailabilityZone:   awssdk.String("eu-west-3"),
						Type:               awssdk.String("gp2"),
						Id:                 "vol-018c5ae89895aca4c",
						Size:               awssdk.Int(23),
						Encrypted:          awssdk.Bool(true),
						MultiAttachEnabled: awssdk.Bool(false),
						Tags: map[string]string{
							"Name": "rootVol",
						},
					},
					&aws.AwsEbsVolume{
						Id: "vol-foobar",
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsEbsVolume{
						Id: "vol-foobar",
					},
					&resource.AbstractResource{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
							"volume_tags": map[string]string{
								"Name": "rootVol",
							},
							"root_block_device": []interface{}{
								map[string]interface{}{
									"volume_id":   "vol-02862d9b39045a3a4",
									"volume_type": "gp2",
									"device_name": "/dev/sda1",
									"encrypted":   true,
									"kms_key_id":  "kms",
									"volume_size": float64(8),
									"iops":        float64(1234),
								},
							},
							"ebs_block_device": []interface{}{
								map[string]interface{}{
									"volume_id":             "vol-018c5ae89895aca4c",
									"volume_type":           "gp2",
									"device_name":           "/dev/sdb",
									"encrypted":             true,
									"delete_on_termination": true,
									"volume_size":           float64(23),
								},
							},
						},
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				factory.On("CreateResource", mock.Anything, "aws_ebs_volume").Times(2).Return(nil, nil)
			},
			false,
		},
	}
	for _, c := range tests {
		t.Run(c.name, func(tt *testing.T) {

			factory := &terraform.MockResourceFactory{}
			if c.mocks != nil {
				c.mocks(factory)
			}

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			aws.InitResourcesMetadata(repo)

			a := NewAwsInstanceBlockDeviceResourceMapper(factory, repo)
			if err := a.Execute(&[]resource.Resource{}, c.args.resourcesFromState); (err != nil) != c.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, c.wantErr)
			}
			changelog, err := diff.Diff(c.args.resourcesFromState, c.args.expectedResource)
			if err != nil {
				tt.Error(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}
