package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
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
					&aws.AwsInstance{
						Id:               "dummy-instance",
						RootBlockDevice:  nil,
						EbsBlockDevice:   nil,
						AvailabilityZone: awssdk.String("eu-west-3"),
						VolumeTags: map[string]string{
							"Name": "rootVol",
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
					&aws.AwsInstance{
						Id:               "dummy-instance",
						AvailabilityZone: awssdk.String("eu-west-3"),
						VolumeTags: map[string]string{
							"Name": "rootVol",
						},
						EbsBlockDevice: &[]struct {
							DeleteOnTermination *bool   `cty:"delete_on_termination"`
							DeviceName          *string `cty:"device_name"`
							Encrypted           *bool   `cty:"encrypted" computed:"true"`
							Iops                *int    `cty:"iops" computed:"true"`
							KmsKeyId            *string `cty:"kms_key_id" computed:"true"`
							SnapshotId          *string `cty:"snapshot_id" computed:"true"`
							VolumeId            *string `cty:"volume_id" computed:"true"`
							VolumeSize          *int    `cty:"volume_size" computed:"true"`
							VolumeType          *string `cty:"volume_type" computed:"true"`
						}{
							{
								DeviceName:          awssdk.String("/dev/sdb"),
								VolumeType:          awssdk.String("gp2"),
								VolumeId:            awssdk.String("vol-018c5ae89895aca4c"),
								VolumeSize:          awssdk.Int(23),
								Encrypted:           awssdk.Bool(true),
								DeleteOnTermination: awssdk.Bool(true),
							},
						},
						RootBlockDevice: &[]struct {
							DeleteOnTermination *bool   `cty:"delete_on_termination"`
							DeviceName          *string `cty:"device_name" computed:"true"`
							Encrypted           *bool   `cty:"encrypted" computed:"true"`
							Iops                *int    `cty:"iops" computed:"true"`
							KmsKeyId            *string `cty:"kms_key_id" computed:"true"`
							VolumeId            *string `cty:"volume_id" computed:"true"`
							VolumeSize          *int    `cty:"volume_size" computed:"true"`
							VolumeType          *string `cty:"volume_type" computed:"true"`
						}{
							{
								DeviceName: awssdk.String("/dev/sda1"),
								Encrypted:  awssdk.Bool(true),
								Iops:       awssdk.Int(1234),
								KmsKeyId:   awssdk.String("kms"),
								VolumeId:   awssdk.String("vol-02862d9b39045a3a4"),
								VolumeSize: awssdk.Int(8),
								VolumeType: awssdk.String("gp2"),
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

			a := NewAwsInstanceBlockDeviceResourceMapper(factory)
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
