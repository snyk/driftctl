package remote

import (
	"testing"

	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/mocks"

	"github.com/snyk/driftctl/test/goldenfile"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestKMSKey(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockKMSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no keys",
			dirName: "aws_kms_key_empty",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllKeys").Return([]*kms.KeyListEntry{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
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
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, "8ee21d91-c000-428c-8032-235aac55da36", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsKmsKeyResourceType, got[0].ResourceType())

				assert.Equal(t, "5d765f32-bfdc-4610-b6ab-f82db5d0601b", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsKmsKeyResourceType, got[1].ResourceType())

				assert.Equal(t, "89d2c023-ea53-40a5-b20a-d84905c622d7", got[2].ResourceId())
				assert.Equal(t, resourceaws.AwsKmsKeyResourceType, got[2].ResourceType())
			},
		},
		{
			test:    "cannot list keys",
			dirName: "aws_kms_key_list",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllKeys").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsKmsKeyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsKmsKeyResourceType, resourceaws.AwsKmsKeyResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

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

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestKMSAlias(t *testing.T) {
	tests := []struct {
		test           string
		dirName        string
		mocks          func(*repository.MockKMSRepository, *mocks.AlerterInterface)
		assertExpected func(*testing.T, []*resource.Resource)
		wantErr        error
	}{
		{
			test:    "no aliases",
			dirName: "aws_kms_alias_empty",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllAliases").Return([]*kms.AliasListEntry{}, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
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
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 3)

				assert.Equal(t, "alias/foo", got[0].ResourceId())
				assert.Equal(t, resourceaws.AwsKmsAliasResourceType, got[0].ResourceType())

				assert.Equal(t, "alias/bar", got[1].ResourceId())
				assert.Equal(t, resourceaws.AwsKmsAliasResourceType, got[1].ResourceType())

				assert.Equal(t, "alias/baz20210225124429210500000001", got[2].ResourceId())
				assert.Equal(t, resourceaws.AwsKmsAliasResourceType, got[2].ResourceType())
			},
		},
		{
			test:    "cannot list aliases",
			dirName: "aws_kms_alias_list",
			mocks: func(repository *repository.MockKMSRepository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllAliases").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsKmsAliasResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsKmsAliasResourceType, resourceaws.AwsKmsAliasResourceType), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
			wantErr: nil,
		},
	}

	factory := terraform.NewTerraformResourceFactory()

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			shouldUpdate := c.dirName == *goldenfile.Update

			sess := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

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

			testFilter := &enumeration.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
