package middlewares

import (
	"reflect"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

func TestAwsInstanceEIP_Execute(t *testing.T) {
	type args struct {
		remoteResources    *[]resource.Resource
		resourcesFromState *[]resource.Resource
	}
	tests := []struct {
		name     string
		args     args
		expected args
	}{
		{
			name: "test that public ip and dns are nulled whith eip",
			args: args{
				remoteResources: &[]resource.Resource{
					&aws.AwsInstance{
						Id:        "instance1",
						PublicIp:  awssdk.String("1.2.3.4"),
						PublicDns: awssdk.String("dns-of-eip.com"),
					},
					&aws.AwsInstance{
						Id:        "instance2",
						PublicIp:  awssdk.String("1.2.3.4"),
						PublicDns: awssdk.String("dns-of-eip.com"),
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsInstance{
						Id:        "instance1",
						PublicIp:  awssdk.String("5.6.7.8"),
						PublicDns: awssdk.String("example.com"),
					},
					&aws.AwsEip{
						Instance: awssdk.String("instance1"),
					},
				},
			},
			expected: args{
				remoteResources: &[]resource.Resource{
					&aws.AwsInstance{
						Id: "instance1",
					},
					&aws.AwsInstance{
						Id:        "instance2",
						PublicIp:  awssdk.String("1.2.3.4"),
						PublicDns: awssdk.String("dns-of-eip.com"),
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsInstance{
						Id: "instance1",
					},
					&aws.AwsEip{
						Instance: awssdk.String("instance1"),
					},
				},
			},
		},
		{
			name: "test that public ip and dns are nulled when eip association",
			args: args{
				remoteResources: &[]resource.Resource{
					&aws.AwsInstance{
						Id:        "instance1",
						PublicIp:  awssdk.String("1.2.3.4"),
						PublicDns: awssdk.String("dns-of-eip.com"),
					},
					&aws.AwsInstance{
						Id:        "instance2",
						PublicIp:  awssdk.String("1.2.3.4"),
						PublicDns: awssdk.String("dns-of-eip.com"),
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsInstance{
						Id:        "instance1",
						PublicIp:  awssdk.String("5.6.7.8"),
						PublicDns: awssdk.String("example.com"),
					},
					&aws.AwsEipAssociation{
						InstanceId: awssdk.String("instance1"),
					},
				},
			},
			expected: args{
				remoteResources: &[]resource.Resource{
					&aws.AwsInstance{
						Id: "instance1",
					},
					&aws.AwsInstance{
						Id:        "instance2",
						PublicIp:  awssdk.String("1.2.3.4"),
						PublicDns: awssdk.String("dns-of-eip.com"),
					},
				},
				resourcesFromState: &[]resource.Resource{
					&aws.AwsInstance{
						Id: "instance1",
					},
					&aws.AwsEipAssociation{
						InstanceId: awssdk.String("instance1"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AwsInstanceEIP{}
			if err := a.Execute(tt.args.remoteResources, tt.args.resourcesFromState); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tt.args, tt.expected) {
				t.Fatalf("Expected results mismatch")
			}
		})
	}
}
