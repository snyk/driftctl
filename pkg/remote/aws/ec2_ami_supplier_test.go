package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"

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
)

func TestEC2AmiSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mock    func(mock *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no amis",
			dirName: "ec2_ami_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllImages").Return([]*ec2.Image{}, nil)
			},
			err: nil,
		},
		{
			test:    "with amis",
			dirName: "ec2_ami_multiple",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllImages").Return([]*ec2.Image{
					{ImageId: aws.String("ami-03a578b46f4c3081b")},
					{ImageId: aws.String("ami-025962fd8b456731f")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list amis",
			dirName: "ec2_ami_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllImages").Return([]*ec2.Image{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsAmiResourceType),
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
			supplierLibrary.AddSupplier(NewEC2AmiSupplier(provider, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockEC2Repository{}
			tt.mock(client)
			s := &EC2AmiSupplier{
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
