package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

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
		test    string
		dirName string
		mocks   func(client *repository.MockRoute53Repository)
		err     error
	}{
		{
			test:    "no zones",
			dirName: "route53_zone_empty",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "single zone",
			dirName: "route53_zone_single",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z08068311RGDXPHF8KE62"),
							Name: awssdk.String("foo.bar"),
						},
					},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "multiples zone (test pagination)",
			dirName: "route53_zone_multiples",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z01809283VH9BBALZHO7B"),
							Name: awssdk.String("foo-0.com"),
						},
						{
							Id:   awssdk.String("Z01804312AV8PHE3C43AD"),
							Name: awssdk.String("foo-1.com"),
						},
						{
							Id:   awssdk.String("Z01874941AR1TCGV5K65C"),
							Name: awssdk.String("foo-2.com"),
						},
					},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "cannot list zones",
			dirName: "route53_zone_empty",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					awserr.NewRequestFailure(nil, 403, ""),
				)
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRoute53ZoneResourceType),
		},
	}
	for _, tt := range tests {
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
			supplierLibrary.AddSupplier(NewRoute53ZoneSupplier(provider, deserializer, repository.NewRoute53Repository(provider.session, cache.New(0))))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockRoute53Repository{}
			tt.mocks(client)
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
