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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2EbsVolumeSupplier_Resources(t *testing.T) {
	tests := []struct {
		test              string
		dirName           string
		volumesPages      mocks.DescribeVolumesPagesOutput
		volumesPagesError error
		err               error
	}{
		{
			test:    "no volumes",
			dirName: "ec2_ebs_volume_empty",
			volumesPages: mocks.DescribeVolumesPagesOutput{
				{
					true,
					&ec2.DescribeVolumesOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "with volumes",
			dirName: "ec2_ebs_volume_multiple",
			volumesPages: mocks.DescribeVolumesPagesOutput{
				{
					false,
					&ec2.DescribeVolumesOutput{
						Volumes: []*ec2.Volume{
							{
								VolumeId: aws.String("vol-081c7272a57a09db1"),
							},
						},
					},
				},
				{
					true,
					&ec2.DescribeVolumesOutput{
						Volumes: []*ec2.Volume{
							{
								VolumeId: aws.String("vol-01ddc91d3d9d1318b"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:              "cannot list volumes",
			dirName:           "ec2_ebs_volume_empty",
			volumesPagesError: awserr.NewRequestFailure(nil, 403, ""),
			err:               remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEbsVolumeResourceType),
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
			supplierLibrary.AddSupplier(NewEC2EbsVolumeSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2EbsVolumeDeserializer()
			client := mocks.NewMockAWSEC2EbsVolumeClient(tt.volumesPages)
			if tt.volumesPagesError != nil {
				client = mocks.NewMockAWSEC2ErrorClient(tt.volumesPagesError)
			}
			s := &EC2EbsVolumeSupplier{
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
