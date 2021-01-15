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
		&aws.AwsSecurityGroup{
			Id:   "sg-test",
			Name: awssdk.String("test"),
		},
		&aws.AwsSecurityGroup{
			Id:   "sg-foo",
			Name: awssdk.String("foo"),
		},
		&aws.AwsDefaultSecurityGroup{
			Id:   "sg-default",
			Name: awssdk.String("default"),
		},
	}
	stateResources := []resource.Resource{
		&aws.AwsSecurityGroup{
			Id:   "sg-bar",
			Name: awssdk.String("bar"),
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
		&aws.AwsSecurityGroup{
			Id:   "sg-test",
			Name: awssdk.String("test"),
		},
		&aws.AwsSecurityGroup{
			Id:   "sg-foo",
			Name: awssdk.String("foo"),
		},
		&aws.AwsDefaultSecurityGroup{
			Id:   "sg-default",
			Name: awssdk.String("default"),
		},
	}
	stateResources := []resource.Resource{
		&aws.AwsDefaultSecurityGroup{
			Id:   "sg-default",
			Name: awssdk.String("default"),
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)
	if err != nil {
		t.Error(err)
	}
	if len(remoteResources) != 3 {
		t.Error("Default security group was ignored")
	}
	managedDefaultSecurityGroup := remoteResources[2].(*aws.AwsDefaultSecurityGroup)
	if *managedDefaultSecurityGroup.Name != "default" {
		t.Error("Default security group is ignored when it should not be")
	}
}
