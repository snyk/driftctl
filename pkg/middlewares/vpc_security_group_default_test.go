package middlewares

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

func TestDefaultVPCSecurityGroupShouldBeIgnored(t *testing.T) {
	middleware := NewVPCDefaultSecurityGroupSanitizer()
	remoteResources := []resource.Resource{
		&resource.AbstractResource{
			Id:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("test"),
			},
		},
		&resource.AbstractResource{
			Id:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("foo"),
			},
		},
		&resource.AbstractResource{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("default"),
			},
		},
	}
	stateResources := []resource.Resource{
		&resource.AbstractResource{
			Id:   "sg-bar",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("bar"),
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(remoteResources) != 2 {
		t.Error("Default security group was not ignored")
	}
}

func TestDefaultVPCSecurityGroupShouldNotBeIgnoredWhenManaged(t *testing.T) {
	middleware := NewVPCDefaultSecurityGroupSanitizer()
	remoteResources := []resource.Resource{
		&resource.AbstractResource{
			Id:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("test"),
			},
		},
		&resource.AbstractResource{
			Id:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("foo"),
			},
		},
		&resource.AbstractResource{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("default"),
			},
		},
	}
	stateResources := []resource.Resource{
		&resource.AbstractResource{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": awssdk.String("default"),
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(remoteResources) != 3 {
		t.Error("Default security group was ignored")
	}
	managedDefaultSecurityGroup := remoteResources[2].(*resource.AbstractResource)
	name, _ := managedDefaultSecurityGroup.Attrs.Get("name")
	if *name.(*string) != "default" {
		t.Error("Default security group is ignored when it should not be")
	}
}
