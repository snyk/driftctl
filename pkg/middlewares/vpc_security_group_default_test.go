package middlewares

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

func TestDefaultVPCSecurityGroupShouldBeIgnored(t *testing.T) {
	middleware := NewVPCDefaultSecurityGroupSanitizer()
	remoteResources := []*resource.Resource{
		{
			Id:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "test",
			},
		},
		{
			Id:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "foo",
			},
		},
		{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	stateResources := []*resource.Resource{
		{
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
	remoteResources := []*resource.Resource{
		{
			Id:   "sg-test",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "test",
			},
		},
		{
			Id:   "sg-foo",
			Type: aws.AwsSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "foo",
			},
		},
		{
			Id:   "sg-default",
			Type: aws.AwsDefaultSecurityGroupResourceType,
			Attrs: &resource.Attributes{
				"name": "default",
			},
		},
	}
	stateResources := []*resource.Resource{
		{
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
	managedDefaultSecurityGroup := remoteResources[2]
	if *managedDefaultSecurityGroup.Attrs.GetString("name") != "default" {
		t.Error("Default security group is ignored when it should not be")
	}
}
