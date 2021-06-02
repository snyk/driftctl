package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testmocks "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func TestKMSKeySupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockKMSRepository)
		err     error
	}{
		{
			test:    "no keys",
			dirName: "kms_key_empty",
			mocks: func(client *repository.MockKMSRepository) {
				client.On("ListAllKeys").Return([]*kms.KeyListEntry{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple keys",
			dirName: "kms_key_multiple",
			mocks: func(client *repository.MockKMSRepository) {
				client.On("ListAllKeys").Return([]*kms.KeyListEntry{
					{KeyId: aws.String("8ee21d91-c000-428c-8032-235aac55da36")},
					{KeyId: aws.String("5d765f32-bfdc-4610-b6ab-f82db5d0601b")},
					{KeyId: aws.String("89d2c023-ea53-40a5-b20a-d84905c622d7")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list keys",
			dirName: "kms_key_empty",
			mocks: func(client *repository.MockKMSRepository) {
				client.On("ListAllKeys").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsKmsKeyResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

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
			supplierLibrary.AddSupplier(NewKMSKeySupplier(provider, deserializer))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := repository.MockKMSRepository{}
			c.mocks(&fakeClient)
			provider := testmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &KMSKeySupplier{
				provider,
				deserializer,
				&fakeClient,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
