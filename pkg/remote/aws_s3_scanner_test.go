package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	tf "github.com/snyk/driftctl/pkg/remote/terraform"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/pkg/remote/aws/client"
	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"

	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestS3Bucket(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockS3Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test: "multiple bucket", dirName: "aws_s3_bucket_multiple",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift2",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift3",
				).Return(
					"ap-northeast-1",
					nil,
				)
			},
		},
		{
			test: "cannot list bucket", dirName: "aws_s3_bucket_list",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllBuckets").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsS3BucketResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketResourceType, resourceaws.AwsS3BucketResourceType), alerts.EnumerationPhase)).Return()
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)
			var repo repository.S3Repository = fakeRepo
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
				repo = repository.NewS3Repository(client.NewAWSClientFactory(session), cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewS3BucketEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsS3BucketResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsS3BucketResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsS3BucketResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestS3BucketInventory(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockS3Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test: "multiple bucket with multiple inventories", dirName: "aws_s3_bucket_inventories_multiple",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift2",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift3",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"ListBucketInventoryConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift2")},
					"eu-west-3",
				).Return(
					[]*s3.InventoryConfiguration{
						{Id: awssdk.String("Inventory_Bucket2")},
						{Id: awssdk.String("Inventory2_Bucket2")},
					},
					nil,
				)
			},
		},
		{
			test: "cannot list bucket", dirName: "aws_s3_bucket_inventories_list_bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllBuckets").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsS3BucketInventoryResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketInventoryResourceType, resourceaws.AwsS3BucketResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
		{
			test: "cannot list bucket inventories", dirName: "aws_s3_bucket_inventories_list_inventories",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllBuckets").Return(
					[]*s3.Bucket{
						{Name: awssdk.String("bucket-martin-test-drift")},
					},
					nil,
				)
				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-3",
					nil,
				)
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On(
					"ListBucketInventoryConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
					"eu-west-3",
				).Return(
					nil,
					awsError,
				)

				alerter.On("SendAlert", resourceaws.AwsS3BucketInventoryResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketInventoryResourceType, resourceaws.AwsS3BucketInventoryResourceType), alerts.EnumerationPhase)).Return()
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)
			var repo repository.S3Repository = fakeRepo
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
				repo = repository.NewS3Repository(client.NewAWSClientFactory(session), cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewS3BucketInventoryEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsS3BucketInventoryResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsS3BucketInventoryResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsS3BucketInventoryResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestS3BucketNotification(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockS3Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "single bucket without notifications",
			dirName: "aws_s3_bucket_notifications_no_notif",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("dritftctl-test-no-notifications")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"dritftctl-test-no-notifications",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketNotification",
					"dritftctl-test-no-notifications",
					"eu-west-3",
				).Return(
					nil,
					nil,
				)
			},
		},
		{
			test: "multiple bucket with notifications", dirName: "aws_s3_bucket_notifications_multiple",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift2",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketNotification",
					"bucket-martin-test-drift2",
					"eu-west-3",
				).Return(
					&s3.NotificationConfiguration{
						LambdaFunctionConfigurations: []*s3.LambdaFunctionConfiguration{
							{
								Id: awssdk.String("tf-s3-lambda-20201103165354926600000001"),
							},
							{
								Id: awssdk.String("tf-s3-lambda-20201103165354926600000002"),
							},
						},
					},
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift3",
				).Return(
					"ap-northeast-1",
					nil,
				)
			},
		},
		{
			test: "Cannot get bucket notification", dirName: "aws_s3_bucket_notifications_list_bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("dritftctl-test-notifications-error")},
				}, nil)
				repository.On(
					"GetBucketLocation",
					"dritftctl-test-notifications-error",
				).Return(
					"eu-west-3",
					nil,
				)
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("GetBucketNotification", "dritftctl-test-notifications-error", "eu-west-3").Return(nil, awsError)

				alerter.On("SendAlert", "aws_s3_bucket_notification.dritftctl-test-notifications-error", alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, "aws_s3_bucket_notification.dritftctl-test-notifications-error", resourceaws.AwsS3BucketNotificationResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
		{
			test: "Cannot list bucket", dirName: "aws_s3_bucket_notifications_list_bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllBuckets").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsS3BucketNotificationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketNotificationResourceType, resourceaws.AwsS3BucketResourceType), alerts.EnumerationPhase)).Return()
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)
			var repo repository.S3Repository = fakeRepo
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
				repo = repository.NewS3Repository(client.NewAWSClientFactory(session), cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewS3BucketNotificationEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsS3BucketNotificationResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsS3BucketNotificationResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsS3BucketNotificationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestS3BucketMetrics(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockS3Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test: "multiple bucket with multiple metrics", dirName: "aws_s3_bucket_metrics_multiple",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift2",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift3",
				).Return(
					"ap-northeast-1",
					nil,
				)

				repository.On(
					"ListBucketMetricsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift2")},
					"eu-west-3",
				).Return(
					[]*s3.MetricsConfiguration{
						{Id: awssdk.String("Metrics_Bucket2")},
						{Id: awssdk.String("Metrics2_Bucket2")},
					},
					nil,
				)
			},
		},
		{
			test: "cannot list bucket", dirName: "aws_s3_bucket_metrics_list_bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllBuckets").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsS3BucketMetricResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketMetricResourceType, resourceaws.AwsS3BucketResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
		{
			test: "cannot list metrics", dirName: "aws_s3_bucket_metrics_list_metrics",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllBuckets").Return(
					[]*s3.Bucket{
						{Name: awssdk.String("bucket-martin-test-drift")},
					},
					nil,
				)
				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-3",
					nil,
				)

				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On(
					"ListBucketMetricsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
					"eu-west-3",
				).Return(
					nil,
					awsError,
				)

				alerter.On("SendAlert", resourceaws.AwsS3BucketMetricResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketMetricResourceType, resourceaws.AwsS3BucketMetricResourceType), alerts.EnumerationPhase)).Return()
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)
			var repo repository.S3Repository = fakeRepo
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
				repo = repository.NewS3Repository(client.NewAWSClientFactory(session), cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewS3BucketMetricsEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsS3BucketMetricResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsS3BucketMetricResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsS3BucketMetricResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestS3BucketPolicy(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockS3Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "single bucket without policy",
			dirName: "aws_s3_bucket_policy_no_policy",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("dritftctl-test-no-policy")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"dritftctl-test-no-policy",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketPolicy",
					"dritftctl-test-no-policy",
					"eu-west-3",
				).Return(
					nil,
					nil,
				)
			},
		},
		{
			test: "multiple bucket with policies", dirName: "aws_s3_bucket_policies_multiple",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift2",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift3",
				).Return(
					"ap-northeast-1",
					nil,
				)

				repository.On(
					"GetBucketPolicy",
					"bucket-martin-test-drift2",
					"eu-west-3",
				).Return(
					// The value here not matter, we only want something not empty
					// to trigger the detail fetcher
					awssdk.String("foobar"),
					nil,
				)

			},
		},
		{
			test: "cannot list bucket", dirName: "aws_s3_bucket_policies_list_bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllBuckets").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsS3BucketPolicyResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketPolicyResourceType, resourceaws.AwsS3BucketResourceType), alerts.EnumerationPhase)).Return()
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)
			var repo repository.S3Repository = fakeRepo
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
				repo = repository.NewS3Repository(client.NewAWSClientFactory(session), cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewS3BucketPolicyEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsS3BucketPolicyResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsS3BucketPolicyResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, err, c.wantErr)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsS3BucketPolicyResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestS3BucketPublicAccessBlock(t *testing.T) {
	dummyError := errors.New("this is an error")

	tests := []struct {
		test           string
		mocks          func(*repository.MockS3Repository, *mocks.AlerterInterface)
		assertExpected func(t *testing.T, got []*resource.Resource)
		wantErr        error
	}{
		{
			test: "multiple bucket, one with access block",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllBuckets").Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-with-public-access-block")},
					{Name: awssdk.String("bucket-without-public-access-block")},
				}, nil)

				repository.On("GetBucketLocation", "bucket-with-public-access-block").
					Return("us-east-1", nil)
				repository.On("GetBucketLocation", "bucket-without-public-access-block").
					Return("us-east-1", nil)

				repository.On("GetBucketPublicAccessBlock", "bucket-with-public-access-block", "us-east-1").
					Return(&s3.PublicAccessBlockConfiguration{
						BlockPublicAcls:   awssdk.Bool(true),
						BlockPublicPolicy: awssdk.Bool(false),
					}, nil)

				repository.On("GetBucketPublicAccessBlock", "bucket-without-public-access-block", "us-east-1").
					Return(nil, nil)
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 1)
				assert.Equal(t, got[0].ResourceId(), "bucket-with-public-access-block")
				assert.Equal(t, got[0].ResourceType(), resourceaws.AwsS3BucketPublicAccessBlockResourceType)
				assert.Equal(t, got[0].Attributes(), &resource.Attributes{
					"block_public_acls":       true,
					"block_public_policy":     false,
					"ignore_public_acls":      false,
					"restrict_public_buckets": false,
				})
			},
		},
		{
			test: "cannot list bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllBuckets").Return(nil, dummyError)
			},
			wantErr: remoteerr.NewResourceListingErrorWithType(dummyError, resourceaws.AwsS3BucketPublicAccessBlockResourceType, resourceaws.AwsS3BucketResourceType),
		},
		{
			test: "cannot list public access block",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllBuckets").Return([]*s3.Bucket{{Name: awssdk.String("foobar")}}, nil)
				repository.On("GetBucketLocation", "foobar").Return("us-east-1", nil)
				repository.On("GetBucketPublicAccessBlock", "foobar", "us-east-1").Return(nil, dummyError)
				alerter.On("SendAlert", "aws_s3_bucket_public_access_block.foobar", alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceScanningError(dummyError, resourceaws.AwsS3BucketPublicAccessBlockResourceType, "foobar"), alerts.EnumerationPhase)).Return()
			},
			assertExpected: func(t *testing.T, got []*resource.Resource) {
				assert.Len(t, got, 0)
			},
		},
	}

	providerVersion := "3.19.0"
	schemaRepository := testresource.InitFakeSchemaRepository("aws", providerVersion)
	resourceaws.InitResourcesMetadata(schemaRepository)
	factory := terraform.NewTerraformResourceFactory(schemaRepository)

	for _, c := range tests {
		t.Run(c.test, func(tt *testing.T) {
			scanOptions := ScannerOptions{}
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.S3Repository = fakeRepo

			remoteLibrary.AddEnumerator(aws.NewS3BucketPublicAccessBlockEnumerator(
				repo, factory,
				tf.TerraformProviderConfig{DefaultAlias: "us-east-1"},
				alerter,
			))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}

			c.assertExpected(tt, got)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
			testFilter.AssertExpectations(tt)
		})
	}
}

func TestS3BucketAnalytic(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockS3Repository, *mocks.AlerterInterface)
		wantErr error
	}{
		{
			test:    "multiple bucket with multiple analytics",
			dirName: "aws_s3_bucket_analytics_multiple",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On(
					"ListAllBuckets",
				).Return([]*s3.Bucket{
					{Name: awssdk.String("bucket-martin-test-drift")},
					{Name: awssdk.String("bucket-martin-test-drift2")},
					{Name: awssdk.String("bucket-martin-test-drift3")},
				}, nil)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-1",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift2",
				).Return(
					"eu-west-3",
					nil,
				)

				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift3",
				).Return(
					"ap-northeast-1",
					nil,
				)

				repository.On(
					"ListBucketAnalyticsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift2")},
					"eu-west-3",
				).Return(
					[]*s3.AnalyticsConfiguration{
						{Id: awssdk.String("Analytics_Bucket2")},
						{Id: awssdk.String("Analytics2_Bucket2")},
					},
					nil,
				)
			},
		},
		{
			test: "cannot list bucket", dirName: "aws_s3_bucket_analytics_list_bucket",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On("ListAllBuckets").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, resourceaws.AwsS3BucketResourceType), alerts.EnumerationPhase)).Return()
			},
			wantErr: nil,
		},
		{
			test: "cannot list Analytics", dirName: "aws_s3_bucket_analytics_list_analytics",
			mocks: func(repository *repository.MockS3Repository, alerter *mocks.AlerterInterface) {
				repository.On("ListAllBuckets").Return(
					[]*s3.Bucket{
						{Name: awssdk.String("bucket-martin-test-drift")},
					},
					nil,
				)
				repository.On(
					"GetBucketLocation",
					"bucket-martin-test-drift",
				).Return(
					"eu-west-3",
					nil,
				)

				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				repository.On(
					"ListBucketAnalyticsConfigurations",
					&s3.Bucket{Name: awssdk.String("bucket-martin-test-drift")},
					"eu-west-3",
				).Return(
					nil,
					awsError,
				)

				alerter.On("SendAlert", resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, resourceaws.AwsS3BucketAnalyticsConfigurationResourceType), alerts.EnumerationPhase)).Return()
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
			alerter := &mocks.AlerterInterface{}
			fakeRepo := &repository.MockS3Repository{}
			c.mocks(fakeRepo, alerter)
			var repo repository.S3Repository = fakeRepo
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
				repo = repository.NewS3Repository(client.NewAWSClientFactory(session), cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewS3BucketAnalyticEnumerator(repo, factory, tf.TerraformProviderConfig{
				Name:         "test",
				DefaultAlias: "eu-west-3",
			}, alerter))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.wantErr, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsS3BucketAnalyticsConfigurationResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
