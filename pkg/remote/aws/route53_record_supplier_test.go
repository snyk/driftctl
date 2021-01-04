package aws

import (
	"context"
	"testing"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestRoute53RecordSupplier_Resources(t *testing.T) {

	tests := []struct {
		test         string
		dirName      string
		zonesPages   mocks.ListHostedZonesPagesOutput
		recordsPages mocks.ListResourceRecordSetsPagesOutput
		err          error
	}{
		{
			test:    "no records",
			dirName: "route53_zone_with_no_record",
			zonesPages: mocks.ListHostedZonesPagesOutput{
				{
					true,
					&route53.ListHostedZonesOutput{
						HostedZones: []*route53.HostedZone{
							{
								Id:   awssdk.String("Z1035360GLIB82T1EH2G"),
								Name: awssdk.String("foo-0.com"),
							},
						},
					},
				},
			},
			recordsPages: mocks.ListResourceRecordSetsPagesOutput{
				{
					true,
					&route53.ListResourceRecordSetsOutput{},
					"Z1035360GLIB82T1EH2G",
				},
			},
			err: nil,
		},
		{
			test:    "multiples records in multiples zones (test pagination)",
			dirName: "route53_record_multiples",
			zonesPages: mocks.ListHostedZonesPagesOutput{
				{
					true,
					&route53.ListHostedZonesOutput{
						HostedZones: []*route53.HostedZone{
							{
								Id:   awssdk.String("Z1035360GLIB82T1EH2G"),
								Name: awssdk.String("foo-0.com"),
							},
							{
								Id:   awssdk.String("Z10347383HV75H96J919W"),
								Name: awssdk.String("foo-1.com"),
							},
						},
					},
				},
			},
			recordsPages: mocks.ListResourceRecordSetsPagesOutput{
				// Zone foo-0.com records below
				{
					false,
					&route53.ListResourceRecordSetsOutput{
						ResourceRecordSets: []*route53.ResourceRecordSet{
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
						},
					},
					"Z1035360GLIB82T1EH2G",
				},
				{
					true,
					&route53.ListResourceRecordSetsOutput{
						ResourceRecordSets: []*route53.ResourceRecordSet{
							{
								Name: awssdk.String("test2"),
								Type: awssdk.String("A"),
							},
							{
								Name: awssdk.String("test3"),
								Type: awssdk.String("A"),
							},
						},
					},
					"Z1035360GLIB82T1EH2G",
				},
				// Zone foo-1.com records below
				{
					true,
					&route53.ListResourceRecordSetsOutput{
						ResourceRecordSets: []*route53.ResourceRecordSet{
							{
								Name: awssdk.String("test2"),
								Type: awssdk.String("A"),
							},
						},
					},
					"Z10347383HV75H96J919W",
				},
			},
			err: nil,
		},
		{
			test:    "explicit subdomain records",
			dirName: "route53_record_explicit_subdomain",
			zonesPages: mocks.ListHostedZonesPagesOutput{
				{
					true,
					&route53.ListHostedZonesOutput{
						HostedZones: []*route53.HostedZone{
							{
								Id:   awssdk.String("Z06486383UC8WYSBZTWFM"),
								Name: awssdk.String("foo-2.com"),
							},
						},
					},
				},
			},
			recordsPages: mocks.ListResourceRecordSetsPagesOutput{
				{
					true,
					&route53.ListResourceRecordSetsOutput{
						ResourceRecordSets: []*route53.ResourceRecordSet{
							{
								Name: awssdk.String("test0"),
								Type: awssdk.String("TXT"),
							},
							{
								Name: awssdk.String("test1.foo-2.com"),
								Type: awssdk.String("TXT"),
							},
							{
								Name: awssdk.String("_test2.foo-2.com"),
								Type: awssdk.String("TXT"),
							},
						},
					},
					"Z06486383UC8WYSBZTWFM",
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			shouldUpdate := tt.dirName == *goldenfile.Update
			if shouldUpdate {
				provider, err := NewTerraFormProvider()
				if err != nil {
					t.Fatal(err)
				}

				terraform.AddProvider(terraform.AWS, provider)
				resource.AddSupplier(NewRoute53RecordSupplier(provider.Runner(), route53.New(provider.session)))
			}

			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewRoute53RecordDeserializer()
			s := &Route53RecordSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSRoute53RecordClient(tt.zonesPages, tt.recordsPages),
				terraform.NewParallelResourceReader(pkg.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
