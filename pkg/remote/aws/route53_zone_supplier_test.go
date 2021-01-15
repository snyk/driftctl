package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
)

func TestRoute53ZoneSupplier_Resources(t *testing.T) {

	tests := []struct {
		test       string
		dirName    string
		zonesPages mocks.ListHostedZonesPagesOutput
		err        error
	}{
		{
			test:    "no zones",
			dirName: "route53_zone_empty",
			zonesPages: mocks.ListHostedZonesPagesOutput{
				{
					true,
					&route53.ListHostedZonesOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "single zone",
			dirName: "route53_zone_single",
			zonesPages: mocks.ListHostedZonesPagesOutput{
				{
					true,
					&route53.ListHostedZonesOutput{
						HostedZones: []*route53.HostedZone{
							{
								Id:   awssdk.String("Z08068311RGDXPHF8KE62"),
								Name: awssdk.String("foo.bar"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:    "multiples zone (test pagination)",
			dirName: "route53_zone_multiples",
			zonesPages: mocks.ListHostedZonesPagesOutput{
				{
					false,
					&route53.ListHostedZonesOutput{
						HostedZones: []*route53.HostedZone{
							{
								Id:   awssdk.String("Z01809283VH9BBALZHO7B"),
								Name: awssdk.String("foo-0.com"),
							},
							{
								Id:   awssdk.String("Z01804312AV8PHE3C43AD"),
								Name: awssdk.String("foo-1.com"),
							},
						},
					},
				},
				{
					true,
					&route53.ListHostedZonesOutput{
						HostedZones: []*route53.HostedZone{
							{
								Id:   awssdk.String("Z01874941AR1TCGV5K65C"),
								Name: awssdk.String("foo-2.com"),
							},
						},
					},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewRoute53ZoneSupplier(provider.Runner(), route53.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			deserializer := awsdeserializer.NewRoute53ZoneDeserializer()
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			s := &Route53ZoneSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSRoute53ZoneClient(tt.zonesPages),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
