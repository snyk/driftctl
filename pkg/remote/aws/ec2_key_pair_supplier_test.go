package aws

import (
	"context"
	"testing"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestEC2KeyPairSupplier_Resources(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		kpNames []string
		err     error
	}{
		{
			test:    "no key pairs",
			dirName: "ec2_key_pair_empty",
			kpNames: []string{},
			err:     nil,
		},
		{
			test:    "with key pairs",
			dirName: "ec2_key_pair_multiple",
			kpNames: []string{"test", "bar"},
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
			resource.AddSupplier(NewEC2KeyPairSupplier(provider.Runner(), ec2.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2KeyPairDeserializer()
			s := &EC2KeyPairSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSEC2KeyPairClient(tt.kpNames),
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

func TestEC2KeyPair_Diff(t *testing.T) {
	tests := []struct {
		test      string
		firstRes  resourceaws.AwsKeyPair
		secondRes resourceaws.AwsKeyPair
		wantErr   bool
	}{
		{
			test: "no diff - identical resource",
			firstRes: resourceaws.AwsKeyPair{
				Id: "foo",
			},
			secondRes: resourceaws.AwsKeyPair{
				Id: "foo",
			},
			wantErr: false,
		},
		{
			test: "no diff - with PublicKey and KeyNamePrefix",
			firstRes: resourceaws.AwsKeyPair{
				Id:            "bar",
				PublicKey:     aws.String("ssh-rsa BBBBB3NzaC1yc2E"),
				KeyNamePrefix: aws.String("test"),
			},
			secondRes: resourceaws.AwsKeyPair{
				Id:            "bar",
				PublicKey:     nil,
				KeyNamePrefix: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			changelog, err := diff.Diff(tt.firstRes, tt.secondRes)
			if err != nil {
				panic(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("got = %v, want %v", awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}
