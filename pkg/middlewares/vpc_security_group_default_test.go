package middlewares

import (
	"testing"

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
				"name": "test",
			},
		},
		&resource.AbstractResource{
			Id:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "foo",
			},
		},
		&resource.AbstractResource{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	stateResources := []resource.Resource{
		&resource.AbstractResource{
			Id:   "sg-bar",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "bar",
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
				"name": "test",
			},
		},
		&resource.AbstractResource{
			Id:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "foo",
			},
		},
		&resource.AbstractResource{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	stateResources := []resource.Resource{
		&resource.AbstractResource{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
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
	if *managedDefaultSecurityGroup.Attrs.GetString("name") != "default" {
		t.Error("Default security group is ignored when it should not be")
	}
}
