package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

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
		listError  error
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
		{
			test:      "cannot list zones",
			dirName:   "route53_zone_empty",
			listError: awserr.NewRequestFailure(nil, 403, ""),
			err:       remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRoute53ZoneResourceType),
		},
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewRoute53ZoneSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			deserializer := awsdeserializer.NewRoute53ZoneDeserializer()
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := mocks.NewMockAWSRoute53ZoneClient(tt.zonesPages, tt.listError)
			s := &Route53ZoneSupplier{
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
