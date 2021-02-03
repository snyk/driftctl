package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

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
		test           string
		dirName        string
		instancesPages mocks.DescribeInstancesPagesOutput
		listError      error
		err            error
	}{
		{
			test:    "no instances",
			dirName: "ec2_instance_empty",
			instancesPages: mocks.DescribeInstancesPagesOutput{
				{
					true,
					&ec2.DescribeInstancesOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "with instances",
			dirName: "ec2_instance_multiple",
			instancesPages: mocks.DescribeInstancesPagesOutput{
				{
					false,
					&ec2.DescribeInstancesOutput{
						Reservations: []*ec2.Reservation{
							{
								Instances: []*ec2.Instance{
									{
										InstanceId: aws.String("i-0d3650a23f4e45dc0"),
									},
								},
							},
						},
					},
				},
				{
					true,
					&ec2.DescribeInstancesOutput{
						Reservations: []*ec2.Reservation{
							{
								Instances: []*ec2.Instance{
									{
										InstanceId: aws.String("i-010376047a71419f1"),
									},
								},
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:    "with terminated instances",
			dirName: "ec2_instance_terminated",
			instancesPages: mocks.DescribeInstancesPagesOutput{
				{
					true,
					&ec2.DescribeInstancesOutput{
						Reservations: []*ec2.Reservation{
							{
								Instances: []*ec2.Instance{
									{
										InstanceId: aws.String("i-0e1543baf4f2cd990"),
									},
									{
										InstanceId: aws.String("i-0a3a7ed51ae2b4fa0"), // Nil
									},
								},
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:      "Cannot list instances",
			dirName:   "ec2_instance_empty",
			listError: awserr.NewRequestFailure(nil, 403, ""),
			err:       remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsInstanceResourceType),
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
			supplierLibrary.AddSupplier(NewEC2InstanceSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2InstanceDeserializer()
			client := mocks.NewMockAWSEC2InstanceClient(tt.instancesPages)
			if tt.listError != nil {
				client = mocks.NewMockAWSEC2ErrorClient(tt.listError)
			}
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
