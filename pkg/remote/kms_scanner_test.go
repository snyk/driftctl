package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
)

func TestKMSKey(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(repository *repository.MockKMSRepository)
		wantErr error
	}{
		{
			test:    "no keys",
			dirName: "aws_kms_key_empty",
			mocks: func(repository *repository.MockKMSRepository) {
				repository.On("ListAllKeys").Return([]*kms.KeyListEntry{}, nil)
			},
		},
		{
			test:    "multiple keys",
			dirName: "aws_kms_key_multiple",
			mocks: func(repository *repository.MockKMSRepository) {
				repository.On("ListAllKeys").Return([]*kms.KeyListEntry{
					{KeyId: awssdk.String("8ee21d91-c000-428c-8032-235aac55da36")},
					{KeyId: awssdk.String("5d765f32-bfdc-4610-b6ab-f82db5d0601b")},
					{KeyId: awssdk.String("89d2c023-ea53-40a5-b20a-d84905c622d7")},
				}, nil)
			},
		},
		{
			test:    "cannot list keys",
			dirName: "aws_kms_key_list",
			mocks: func(repository *repository.MockKMSRepository) {
				repository.On("ListAllKeys").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			wantErr: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsKmsKeyResourceType),
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
			fakeRepo := &repository.MockKMSRepository{}
			c.mocks(fakeRepo)
			var repo repository.KMSRepository = fakeRepo
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
				repo = repository.NewKMSRepository(sess, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewKMSKeyEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsKmsKeyResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsKmsKeyResourceType, provider, deserializer))

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsKmsKeyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
