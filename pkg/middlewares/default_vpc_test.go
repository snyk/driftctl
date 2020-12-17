package middlewares

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestAwsDefaultVPCShouldBeIgnored(t *testing.T) {
	middleware := NewAwsDefaultVPC()
	remoteResources := []resource.Resource{
		&aws.AwsDefaultVpc{
			Id: "foobar",
		},
	}
	stateResources := []resource.Resource{}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error(err)
	}

	if len(remoteResources) != 0 {
		t.Error("Default VPC was not ignored")
	}
}

func TestAwsDefaultVPCShouldNotBeIgnoredWhenManaged(t *testing.T) {
	middleware := NewAwsDefaultVPC()
	remoteResources := []resource.Resource{
		&aws.AwsDefaultVpc{
			Id: "foobar",
		},
	}
	stateResources := []resource.Resource{
		&aws.AwsDefaultVpc{
			Id: "foobar",
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error(err)
	}

	if len(remoteResources) != 1 {
		t.Error("Default VPC was ignored")
	}
}
