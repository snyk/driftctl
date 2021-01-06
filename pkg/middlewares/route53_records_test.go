package middlewares

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

func TestDefaultRoute53RecordShouldBeIgnored(t *testing.T) {
	middleware := NewRoute53DefaultZoneRecordSanitizer()
	remoteResources := []resource.Resource{
		&aws.AwsRoute53Zone{},
		&aws.AwsRoute53Record{
			Type: awssdk.String("NS"),
			Id:   "123_foobar_NS",
		},
		&aws.AwsRoute53Record{
			Type: awssdk.String("SOA"),
			Id:   "123_foobar_SOA",
		},
		&aws.AwsRoute53Record{
			Type: awssdk.String("A"),
			Id:   "123_foobar_A",
		},
	}
	stateResources := []resource.Resource{
		&aws.AwsRoute53Record{
			Type: awssdk.String("NS"),
			Id:   "456_barfoo_NS",
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error(err)
	}

	if len(remoteResources) != 2 {
		t.Error("Default records were not ignored")
	}
	remainingResource := remoteResources[1].(*aws.AwsRoute53Record)
	if *remainingResource.Type != "A" {
		t.Error("Default record is invalid")
	}
}

func TestDefaultRoute53RecordShouldNotBeIgnoredWhenManaged(t *testing.T) {
	middleware := NewRoute53DefaultZoneRecordSanitizer()
	remoteResources := []resource.Resource{
		&aws.AwsRoute53Zone{},
		&aws.AwsRoute53Record{
			Type: awssdk.String("NS"),
			Id:   "123_foobar_NS",
		},
		&aws.AwsRoute53Record{
			Type: awssdk.String("SOA"),
			Id:   "123_foobar_SOA",
		},
		&aws.AwsRoute53Record{
			Type: awssdk.String("A"),
			Id:   "123_foobar_A",
		},
	}
	stateResources := []resource.Resource{
		&aws.AwsRoute53Record{
			Type: awssdk.String("NS"),
			Id:   "123_foobar_NS",
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error(err)
	}

	if len(remoteResources) != 3 {
		t.Error("Default records were not ignored")
	}
	managedDefaultRecord := remoteResources[1].(*aws.AwsRoute53Record)
	if *managedDefaultRecord.Type != "NS" {
		t.Error("Default record is ignored but sholuld not be")
	}

	ignoredRecord := remoteResources[2].(*aws.AwsRoute53Record)
	if *ignoredRecord.Type != "A" {
		t.Error("Non default record was ignored")
	}
}
