package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	tf "github.com/cloudskiff/driftctl/pkg/remote/terraform"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
)

func TestEC2EbsVolume(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockEC2Repository)
		wantErr error
	}{
		{
			test:    "no volumes",
			dirName: "aws_ec2_ebs_volume_empty",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllVolumes").Return([]*ec2.Volume{}, nil)
			},
		},
		{
			test:    "multiple volumes",
			dirName: "aws_ec2_ebs_volume_multiple",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllVolumes").Return([]*ec2.Volume{
					{VolumeId: awssdk.String("vol-081c7272a57a09db1")},
					{VolumeId: awssdk.String("vol-01ddc91d3d9d1318b")},
				}, nil)
			},
		},
		{
			test:    "cannot list volumes",
			dirName: "aws_ec2_ebs_volume_list",
			mocks: func(repository *repository.MockEC2Repository) {
				repository.On("ListAllVolumes").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsEbsVolumeResourceType),
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)
	alerter := &mocks.AlerterInterface{}

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			fakeRepo := &repository.MockEC2Repository{}
			c.mocks(fakeRepo)
			var repo repository.EC2Repository = fakeRepo
			providerVersion := "3.19.0"
			realProvider, err := terraform2.InitTestAwsProvider(providerLibrary, providerVersion)
			if err != nil {
				t.Fatal(err)
			}
			provider := terraform2.NewFakeTerraformProvider(realProvider)
			provider.WithResponse(c.dirName)

			// Replace mock by real resources if we are in update mode
			if shouldUpdate {
				err := realProvider.Init()
				if err != nil {
					t.Fatal(err)
				}
				provider.ShouldUpdate()
				repo = repository.NewEC2Repository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewEC2EbsVolumeEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsEbsVolumeResourceType, common.NewGenericDetailFetcher(resourceaws.AwsEbsVolumeResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsEbsVolumeResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
