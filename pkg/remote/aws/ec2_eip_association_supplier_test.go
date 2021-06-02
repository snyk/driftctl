package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestEC2EipAssociationSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mock    func(mock *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no eip associations",
			dirName: "ec2_eip_association_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllAddressesAssociation").Return([]string{}, nil)
			},
			err: nil,
		},
		{
			test:    "with eip associations",
			dirName: "ec2_eip_association_single",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllAddressesAssociation").Return([]string{
					"eipassoc-0e9a7356e30f0c3d1",
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list eip associations",
			dirName: "ec2_eip_association_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllAddressesAssociation").Return([]string{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEipAssociationResourceType),
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
			supplierLibrary.AddSupplier(NewEC2EipAssociationSupplier(provider, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockEC2Repository{}
			tt.mock(client)
			s := &EC2EipAssociationSupplier{
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
