package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/parallel"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/r3labs/diff/v2"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestEC2KeyPairSupplier_Resources(t *testing.T) {
	tests := []struct {
		test      string
		dirName   string
		kpNames   []string
		listError error
		err       error
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
		{
			test:      "cannot list key pairs",
			dirName:   "ec2_key_pair_empty",
			kpNames:   []string{},
			listError: awserr.NewRequestFailure(nil, 403, ""),
			err:       remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsKeyPairResourceType),
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
			supplierLibrary.AddSupplier(NewEC2KeyPairSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewEC2KeyPairDeserializer()
			client := mocks.NewMockAWSEC2KeyPairClient(tt.kpNames)
			if tt.listError != nil {
				client = mocks.NewMockAWSEC2ErrorClient(tt.listError)
			}
			s := &EC2KeyPairSupplier{
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
