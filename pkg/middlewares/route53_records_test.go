package middlewares

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource/aws"

	"github.com/snyk/driftctl/enumeration/resource"
)

func TestDefaultRoute53RecordShouldBeIgnored(t *testing.T) {
	middleware := NewRoute53DefaultZoneRecordSanitizer()
	remoteResources := []*resource.Resource{
		{
			Type:  aws.AwsRoute53ZoneResourceType,
			Attrs: &resource.Attributes{},
		},
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_NS",
			Attrs: &resource.Attributes{
				"type": "NS",
			},
		},
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_SOA",
			Attrs: &resource.Attributes{
				"type": "SOA",
			},
		},
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_A",
			Attrs: &resource.Attributes{
				"type": "A",
			},
		},
	}
	stateResources := []*resource.Resource{
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "456_barfoo_NS",
			Attrs: &resource.Attributes{
				"type": "NS",
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error(err)
	}

	if len(remoteResources) != 2 {
		t.Error("Default records were not ignored")
	}
	remainingResource := remoteResources[1]
	ty, _ := remainingResource.Attrs.Get("type")
	if ty != "A" {
		t.Error("Default record is invalid")
	}
}

func TestDefaultRoute53RecordShouldNotBeIgnoredWhenManaged(t *testing.T) {
	middleware := NewRoute53DefaultZoneRecordSanitizer()
	remoteResources := []*resource.Resource{
		{
			Type:  aws.AwsRoute53ZoneResourceType,
			Attrs: &resource.Attributes{},
		},
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_NS",
			Attrs: &resource.Attributes{
				"type": "NS",
			},
		},
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_SOA",
			Attrs: &resource.Attributes{
				"type": "SOA",
			},
		},
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_A",
			Attrs: &resource.Attributes{
				"type": "A",
			},
		},
	}
	stateResources := []*resource.Resource{
		{
			Type: aws.AwsRoute53RecordResourceType,
			Id:   "123_foobar_NS",
			Attrs: &resource.Attributes{
				"type": "NS",
			},
		},
	}
	err := middleware.Execute(&remoteResources, &stateResources)

	if err != nil {
		t.Error(err)
	}

	if len(remoteResources) != 3 {
		t.Error("Default records were not ignored")
	}
	managedDefaultRecord := remoteResources[1]
	ty, _ := managedDefaultRecord.Attrs.Get("type")
	if ty != "NS" {
		t.Error("Default record is ignored but should not be")
	}

	ignoredRecord := remoteResources[2]
	ty, _ = ignoredRecord.Attrs.Get("type")
	if ty != "A" {
		t.Error("Non default record was ignored")
	}
}
