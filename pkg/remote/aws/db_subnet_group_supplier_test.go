package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestDBSubnetGroupSupplier_Resources(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockRDSRepository)
		err     error
	}{
		{
			test:    "no subnets",
			dirName: "db_subnet_empty",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDbSubnetGroups").Return([]*rds.DBSubnetGroup{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiples db subnets",
			dirName: "db_subnet_multiples",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDbSubnetGroups").Return([]*rds.DBSubnetGroup{
					{
						DBSubnetGroupName: aws.String("foo"),
					},
					{
						DBSubnetGroupName: aws.String("bar"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list subnet",
			dirName: "db_subnet_empty",
			mocks: func(client *repository.MockRDSRepository) {
				client.On("ListAllDbSubnetGroups").Return([]*rds.DBSubnetGroup{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsDbSubnetGroupResourceType),
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
			supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockRDSRepository{}
			tt.mocks(client)
			s := &DBSubnetGroupSupplier{
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
