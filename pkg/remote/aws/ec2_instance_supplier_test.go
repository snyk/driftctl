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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2InstanceSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mock    func(mock *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no instances",
			dirName: "ec2_instance_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllInstances").Return([]*ec2.Instance{}, nil)
			},
			err: nil,
		},
		{
			test:    "with instances",
			dirName: "ec2_instance_multiple",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllInstances").Return([]*ec2.Instance{
					{
						InstanceId: aws.String("i-0d3650a23f4e45dc0"),
					},
					{
						InstanceId: aws.String("i-010376047a71419f1"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "with terminated instances",
			dirName: "ec2_instance_terminated",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllInstances").Return([]*ec2.Instance{
					{
						InstanceId: aws.String("i-0e1543baf4f2cd990"),
					},
					{
						InstanceId: aws.String("i-0a3a7ed51ae2b4fa0"), // Nil
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "Cannot list instances",
			dirName: "ec2_instance_empty",
			mock: func(mock *repository.MockEC2Repository) {
				mock.On("ListAllInstances").Return([]*ec2.Instance{}, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsInstanceResourceType),
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
			supplierLibrary.AddSupplier(NewEC2InstanceSupplier(provider, deserializer))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			client := &repository.MockEC2Repository{}
			tt.mock(client)
			s := &EC2InstanceSupplier{
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
