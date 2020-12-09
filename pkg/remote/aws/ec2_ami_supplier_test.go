package aws

import (
	"context"
	"testing"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2AmiSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		amiIDs  []string
		err     error
	}{
		{
			test:    "no amis",
			dirName: "ec2_ami_empty",
			amiIDs:  []string{},
			err:     nil,
		},
		{
			test:    "with amis",
			dirName: "ec2_ami_multiple",
			amiIDs:  []string{"ami-03a578b46f4c3081b", "ami-025962fd8b456731f"},
			err:     nil,
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
			resource.AddSupplier(NewEC2AmiSupplier(provider.Runner(), ec2.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2AmiDeserializer()
			s := &EC2AmiSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSEC2AmiClient(tt.amiIDs),
				terraform.NewParallelResourceReader(pkg.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
