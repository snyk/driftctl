package middlewares

import (
	"github.com/snyk/driftctl/enumeration/terraform"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/stretchr/testify/mock"
)

func TestAwsInstanceBlockDeviceResourceMapper_Execute(t *testing.T) {
	type args struct {
		expectedResource   *[]*resource.Resource
		resourcesFromState *[]*resource.Resource
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
				expectedResource   *[]*resource.Resource
				resourcesFromState *[]*resource.Resource
			}{
				expectedResource: &[]*resource.Resource{
					{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
						},
					},
					{
						Id:   "vol-02862d9b39045a3a4",
						Type: "aws_ebs_volume",
						Attrs: &resource.Attributes{
							"id":                   "vol-02862d9b39045a3a4",
							"encrypted":            true,
							"multi_attach_enabled": false,
							"availability_zone":    "eu-west-3",
							"iops":                 1234,
							"kms_key_id":           "kms",
							"size":                 8,
							"type":                 "gp2",
							"throughput":           125,
							"tags": map[string]interface{}{
								"Name": "rootVol",
							},
						},
					},
					{
						Id:   "vol-018c5ae89895aca4c",
						Type: "aws_ebs_volume",
						Attrs: &resource.Attributes{
							"id":                   "vol-018c5ae89895aca4c",
							"encrypted":            true,
							"multi_attach_enabled": false,
							"availability_zone":    "eu-west-3",
							"size":                 23,
							"type":                 "gp2",
							"throughput":           125,
							"tags": map[string]interface{}{
								"Name": "rootVol",
							},
						},
					},
					{
						Id:    "vol-foobar",
						Type:  "aws_ebs_volume",
						Attrs: &resource.Attributes{},
					},
				},
				resourcesFromState: &[]*resource.Resource{
					{
						Id:    "vol-foobar",
						Type:  "aws_ebs_volume",
						Attrs: &resource.Attributes{},
					},
					{
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
									"volume_size": 8,
									"throughput":  125,
									"iops":        1234,
								},
							},
							"ebs_block_device": []interface{}{
								map[string]interface{}{
									"volume_id":             "vol-018c5ae89895aca4c",
									"volume_type":           "gp2",
									"device_name":           "/dev/sdb",
									"encrypted":             true,
									"delete_on_termination": true,
									"volume_size":           23,
									"throughput":            125,
								},
							},
						},
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				foo := resource.Resource{
					Id:   "vol-02862d9b39045a3a4",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"id":                   "vol-02862d9b39045a3a4",
						"encrypted":            true,
						"multi_attach_enabled": false,
						"availability_zone":    "eu-west-3",
						"iops":                 1234,
						"kms_key_id":           "kms",
						"size":                 8,
						"type":                 "gp2",
						"throughput":           125,
						"tags": map[string]interface{}{
							"Name": "rootVol",
						},
					},
				}
				factory.On("CreateAbstractResource", "aws_ebs_volume", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "vol-02862d9b39045a3a4"
				})).Times(1).Return(&foo, nil)

				bar := resource.Resource{
					Id:   "vol-018c5ae89895aca4c",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"id":                   "vol-018c5ae89895aca4c",
						"encrypted":            true,
						"multi_attach_enabled": false,
						"availability_zone":    "eu-west-3",
						"size":                 23,
						"type":                 "gp2",
						"throughput":           125,
						"tags": map[string]interface{}{
							"Name": "rootVol",
						},
					},
				}
				factory.On("CreateAbstractResource", "aws_ebs_volume", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "vol-018c5ae89895aca4c"
				})).Times(1).Return(&bar, nil)
			},
			false,
		},
		{
			"Test with tags inside root/ebs block device",
			struct {
				expectedResource   *[]*resource.Resource
				resourcesFromState *[]*resource.Resource
			}{
				expectedResource: &[]*resource.Resource{
					&resource.Resource{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
						},
					},
					&resource.Resource{
						Id:   "vol-02862d9b39045a3a4",
						Type: "aws_ebs_volume",
						Attrs: &resource.Attributes{
							"id":                   "vol-02862d9b39045a3a4",
							"encrypted":            true,
							"multi_attach_enabled": false,
							"availability_zone":    "eu-west-3",
							"iops":                 1234,
							"kms_key_id":           "kms",
							"size":                 8,
							"type":                 "gp2",
							"throughput":           125,
							"tags": map[string]interface{}{
								"Name": "rootVol",
							},
						},
					},
					&resource.Resource{
						Id:   "vol-018c5ae89895aca4c",
						Type: "aws_ebs_volume",
						Attrs: &resource.Attributes{
							"id":                   "vol-018c5ae89895aca4c",
							"encrypted":            true,
							"multi_attach_enabled": false,
							"availability_zone":    "eu-west-3",
							"size":                 23,
							"type":                 "gp2",
							"throughput":           125,
							"tags": map[string]interface{}{
								"Name": "ebsVol",
							},
						},
					},
				},
				resourcesFromState: &[]*resource.Resource{
					&resource.Resource{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
							"root_block_device": []interface{}{
								map[string]interface{}{
									"volume_id":   "vol-02862d9b39045a3a4",
									"volume_type": "gp2",
									"device_name": "/dev/sda1",
									"encrypted":   true,
									"kms_key_id":  "kms",
									"volume_size": 8,
									"throughput":  125,
									"iops":        1234,
									"tags": map[string]interface{}{
										"Name": "rootVol",
									},
								},
							},
							"ebs_block_device": []interface{}{
								map[string]interface{}{
									"volume_id":             "vol-018c5ae89895aca4c",
									"volume_type":           "gp2",
									"device_name":           "/dev/sdb",
									"encrypted":             true,
									"delete_on_termination": true,
									"volume_size":           23,
									"throughput":            125,
									"tags": map[string]interface{}{
										"Name": "ebsVol",
									},
								},
							},
						},
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {
				foo := resource.Resource{
					Id:   "vol-02862d9b39045a3a4",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"id":                   "vol-02862d9b39045a3a4",
						"encrypted":            true,
						"multi_attach_enabled": false,
						"availability_zone":    "eu-west-3",
						"iops":                 1234,
						"kms_key_id":           "kms",
						"size":                 8,
						"type":                 "gp2",
						"throughput":           125,
						"tags": map[string]interface{}{
							"Name": "rootVol",
						},
					},
				}
				factory.On("CreateAbstractResource", "aws_ebs_volume", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "vol-02862d9b39045a3a4" && len(input["tags"].(map[string]interface{})) == 1
				})).Times(1).Return(&foo, nil)

				bar := resource.Resource{
					Id:   "vol-018c5ae89895aca4c",
					Type: "aws_ebs_volume",
					Attrs: &resource.Attributes{
						"id":                   "vol-018c5ae89895aca4c",
						"encrypted":            true,
						"multi_attach_enabled": false,
						"availability_zone":    "eu-west-3",
						"size":                 23,
						"type":                 "gp2",
						"throughput":           125,
						"tags": map[string]interface{}{
							"Name": "ebsVol",
						},
					},
				}
				factory.On("CreateAbstractResource", "aws_ebs_volume", mock.Anything, mock.MatchedBy(func(input map[string]interface{}) bool {
					return input["id"] == "vol-018c5ae89895aca4c" && len(input["tags"].(map[string]interface{})) == 1
				})).Times(1).Return(&bar, nil)
			},
			false,
		},
		{
			"Should not create ebs volume if there is already one (e.g. inline ebs_block_device)",
			struct {
				expectedResource   *[]*resource.Resource
				resourcesFromState *[]*resource.Resource
			}{
				expectedResource: &[]*resource.Resource{
					&resource.Resource{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
						},
					},
					&resource.Resource{
						Id:   "vol-02862d9b39045a3a4",
						Type: "aws_ebs_volume",
						Attrs: &resource.Attributes{
							"id": "vol-02862d9b39045a3a4",
						},
					},
				},
				resourcesFromState: &[]*resource.Resource{
					&resource.Resource{
						Id:   "vol-02862d9b39045a3a4",
						Type: "aws_ebs_volume",
						Attrs: &resource.Attributes{
							"id": "vol-02862d9b39045a3a4",
						},
					},
					&resource.Resource{
						Id:   "dummy-instance",
						Type: "aws_instance",
						Attrs: &resource.Attributes{
							"availability_zone": "eu-west-3",
							"ebs_block_device": []interface{}{
								map[string]interface{}{
									"volume_id": "vol-02862d9b39045a3a4",
								},
							},
						},
					},
				},
			},
			func(factory *terraform.MockResourceFactory) {},
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
			if err := a.Execute(&[]*resource.Resource{}, c.args.resourcesFromState); (err != nil) != c.wantErr {
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
