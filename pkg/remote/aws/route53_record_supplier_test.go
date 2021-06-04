package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	testmocks "github.com/cloudskiff/driftctl/test/mocks"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
)

func TestRoute53RecordSupplier_Resources(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockRoute53Repository)
		err     error
	}{
		{
			test:    "no records",
			dirName: "route53_zone_with_no_record",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z1035360GLIB82T1EH2G"),
							Name: awssdk.String("foo-0.com"),
						},
					},
					nil,
				)
				client.On("ListRecordsForZone", "Z1035360GLIB82T1EH2G").Return([]*route53.ResourceRecordSet{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiples records in multiples zones",
			dirName: "route53_record_multiples",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z1035360GLIB82T1EH2G"),
							Name: awssdk.String("foo-0.com"),
						},
						{
							Id:   awssdk.String("Z10347383HV75H96J919W"),
							Name: awssdk.String("foo-1.com"),
						},
					},
					nil,
				)
				client.On("ListRecordsForZone", "Z1035360GLIB82T1EH2G").Return([]*route53.ResourceRecordSet{
					{
						Name: awssdk.String("foo-0.com"),
						Type: awssdk.String("NS"),
					},
					{
						Name: awssdk.String("test0"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test1"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test2"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test3"),
						Type: awssdk.String("A"),
					},
				}, nil)
				client.On("ListRecordsForZone", "Z10347383HV75H96J919W").Return([]*route53.ResourceRecordSet{
					{
						Name: awssdk.String("test2"),
						Type: awssdk.String("A"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "explicit subdomain records",
			dirName: "route53_record_explicit_subdomain",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z06486383UC8WYSBZTWFM"),
							Name: awssdk.String("foo-2.com"),
						},
					},
					nil,
				)
				client.On("ListRecordsForZone", "Z06486383UC8WYSBZTWFM").Return([]*route53.ResourceRecordSet{
					{
						Name: awssdk.String("test0"),
						Type: awssdk.String("TXT"),
					},
					{
						Name: awssdk.String("test0"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test1.foo-2.com"),
						Type: awssdk.String("TXT"),
					},
					{
						Name: awssdk.String("test1.foo-2.com"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("_test2.foo-2.com"),
						Type: awssdk.String("TXT"),
					},
					{
						Name: awssdk.String("_test2.foo-2.com"),
						Type: awssdk.String("A"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list zones",
			dirName: "route53_zone_with_no_record",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationErrorWithType(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRoute53RecordResourceType, resourceaws.AwsRoute53ZoneResourceType),
		},
		{
			test:    "cannot list records",
			dirName: "route53_zone_with_no_record",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z06486383UC8WYSBZTWFM"),
							Name: awssdk.String("foo-2.com"),
						},
					},
					nil)
				client.On("ListRecordsForZone", "Z06486383UC8WYSBZTWFM").Return(
					[]*route53.ResourceRecordSet{},
					awserr.NewRequestFailure(nil, 403, ""))

			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRoute53RecordResourceType),
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			shouldUpdate := tt.dirName == *goldenfile.Update

			providerLibrary := terraform.NewProviderLibrary()
			supplierLibrary := resource.NewSupplierLibrary()

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			resourceaws.InitResourcesMetadata(repo)
			factory := terraform.NewTerraformResourceFactory(repo)

			deserializer := resource.NewDeserializer(factory)
			if shouldUpdate {
				provider, err := InitTestAwsProvider(providerLibrary)
				if err != nil {
					t.Fatal(err)
				}
				supplierLibrary.AddSupplier(NewRoute53RecordSupplier(provider, deserializer, repository.NewRoute53Repository(provider.session, cache.New(0))))
			}

			provider := testmocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockRoute53Repository{}
			tt.mocks(client)
			s := &Route53RecordSupplier{
				provider,
				deserializer,
				client,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, tt.err, err)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
