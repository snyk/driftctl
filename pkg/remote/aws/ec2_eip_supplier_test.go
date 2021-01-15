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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2EipSupplier_Resources(t *testing.T) {
	tests := []struct {
		test      string
		dirName   string
		addresses []*ec2.Address
		err       error
	}{
		{
			test:      "no eips",
			dirName:   "ec2_eip_empty",
			addresses: []*ec2.Address{},
			err:       nil,
		},
		{
			test:    "with eips",
			dirName: "ec2_eip_multiple",
			addresses: []*ec2.Address{
				{
					AllocationId: aws.String("eipalloc-017d5267e4dda73f1"),
				},
				{
					AllocationId: aws.String("eipalloc-0cf714dc097c992cc"),
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
			resource.AddSupplier(NewEC2EipSupplier(provider.Runner(), ec2.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2EipDeserializer()
			s := &EC2EipSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSEC2EipClient(tt.addresses),
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
