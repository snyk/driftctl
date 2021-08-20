package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKMSKey(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockKMSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no keys",
			dirName: "aws_kms_key_empty",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllKeys").Return([]*kms.KeyListEntry{}, nil)
			},
		},
		{
			test:    "multiple keys",
			dirName: "aws_kms_key_multiple",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
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
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllKeys").Return(nil, awserr.NewRequestFailure(nil, 403, ""))

				alerter.On("SendAlert", resourceaws.AwsKmsKeyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, resourceaws.AwsKmsKeyResourceType, resourceaws.AwsKmsKeyResourceType, alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

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
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockKMSRepository{}
			c.mocks(fakeRepo, alerter)

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

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsKmsKeyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestKMSAlias(t *testing.T) {
	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockKMSRepository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "no aliases",
			dirName: "aws_kms_alias_empty",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAliases").Return([]*kms.AliasListEntry{}, nil)
			},
		},
		{
			test:    "multiple aliases",
			dirName: "aws_kms_alias_multiple",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAliases").Return([]*kms.AliasListEntry{
					{AliasName: awssdk.String("alias/foo")},
					{AliasName: awssdk.String("alias/bar")},
					{AliasName: awssdk.String("alias/baz20210225124429210500000001")},
				}, nil)
			},
		},
		{
			test:    "cannot list aliases",
			dirName: "aws_kms_alias_list",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAliases").Return(nil, awserr.NewRequestFailure(nil, 403, ""))

				alerter.On("SendAlert", resourceaws.AwsKmsAliasResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, resourceaws.AwsKmsAliasResourceType, resourceaws.AwsKmsAliasResourceType, alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)
	deserializer := resource.NewDeserializer(factory)

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
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockKMSRepository{}
			c.mocks(fakeRepo, alerter)

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

			remoteLibrary.AddEnumerator(aws.NewKMSAliasEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsKmsAliasResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsKmsAliasResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsKmsAliasResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
