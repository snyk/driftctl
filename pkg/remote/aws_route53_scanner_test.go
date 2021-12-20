package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/pkg/errors"
	"github.com/snyk/driftctl/mocks"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/alerts"
	"github.com/snyk/driftctl/pkg/remote/aws"
	"github.com/snyk/driftctl/pkg/remote/cache"
	"github.com/snyk/driftctl/pkg/remote/common"
	remoteerr "github.com/snyk/driftctl/pkg/remote/error"
	testresource "github.com/snyk/driftctl/test/resource"
	terraform2 "github.com/snyk/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/pkg/resource"
	resourceaws "github.com/snyk/driftctl/pkg/resource/aws"

	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestRoute53_HealthCheck(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockRoute53Repository, *mocks.AlerterInterface)
		err     error
	}{
		{
			test:    "no health check",
			dirName: "aws_route53_health_check_empty",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllHealthChecks").Return([]*route53.HealthCheck{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple health check",
			dirName: "aws_route53_health_check_multiple",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllHealthChecks").Return([]*route53.HealthCheck{
					{Id: awssdk.String("7001a9df-ded4-4802-9909-668eb80b972b")},
					{Id: awssdk.String("84fc318a-2e0d-41d6-b638-280e2f0f4e26")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list health check",
			dirName: "aws_route53_health_check_empty",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllHealthChecks").Return(nil, awsError)

				alerter.On("SendAlert", resourceaws.AwsRoute53HealthCheckResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRoute53HealthCheckResourceType, resourceaws.AwsRoute53HealthCheckResourceType), alerts.EnumerationPhase)).Return()
			},
			err: nil,
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
			fakeRepo := &repository.MockRoute53Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.Route53Repository = fakeRepo
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
				repo = repository.NewRoute53Repository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewRoute53HealthCheckEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRoute53HealthCheckResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsRoute53HealthCheckResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRoute53HealthCheckResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestRoute53_Zone(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockRoute53Repository, *mocks.AlerterInterface)
		err     error
	}{
		{
			test:    "no zones",
			dirName: "aws_route53_zone_empty",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "single zone",
			dirName: "aws_route53_zone_single",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z08068311RGDXPHF8KE62"),
							Name: awssdk.String("foo.bar"),
						},
					},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "multiples zone (test pagination)",
			dirName: "aws_route53_zone_multiples",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z01809283VH9BBALZHO7B"),
							Name: awssdk.String("foo-0.com"),
						},
						{
							Id:   awssdk.String("Z01804312AV8PHE3C43AD"),
							Name: awssdk.String("foo-1.com"),
						},
						{
							Id:   awssdk.String("Z01874941AR1TCGV5K65C"),
							Name: awssdk.String("foo-2.com"),
						},
					},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "cannot list zones",
			dirName: "aws_route53_zone_empty",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					awsError,
				)

				alerter.On("SendAlert", resourceaws.AwsRoute53ZoneResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRoute53ZoneResourceType, resourceaws.AwsRoute53ZoneResourceType), alerts.EnumerationPhase)).Return()
			},
			err: nil,
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
			fakeRepo := &repository.MockRoute53Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.Route53Repository = fakeRepo
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
				repo = repository.NewRoute53Repository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewRoute53ZoneEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRoute53ZoneResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsRoute53ZoneResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRoute53ZoneResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}

func TestRoute53_Record(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockRoute53Repository, *mocks.AlerterInterface)
		err     error
	}{
		{
			test:    "no records",
			dirName: "aws_route53_zone_with_no_record",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z1035360GLIB82T1EH2G"),
							Name: awssdk.String("foo-0.com"),
						},
					},
					nil,
				)
				client.On("ListRecordsForZone", "Z1035360GLIB82T1EH2G").Return([]*route53.ResourceRecordSet{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiples records in multiples zones",
			dirName: "aws_route53_record_multiples",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z1035360GLIB82T1EH2G"),
							Name: awssdk.String("foo-0.com"),
						},
						{
							Id:   awssdk.String("Z10347383HV75H96J919W"),
							Name: awssdk.String("foo-1.com"),
						},
					},
					nil,
				)
				client.On("ListRecordsForZone", "Z1035360GLIB82T1EH2G").Return([]*route53.ResourceRecordSet{
					{
						Name: awssdk.String("foo-0.com"),
						Type: awssdk.String("NS"),
					},
					{
						Name: awssdk.String("test0"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test1"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test2"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test3"),
						Type: awssdk.String("A"),
					},
				}, nil)
				client.On("ListRecordsForZone", "Z10347383HV75H96J919W").Return([]*route53.ResourceRecordSet{
					{
						Name: awssdk.String("test2"),
						Type: awssdk.String("A"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "explicit subdomain records",
			dirName: "aws_route53_record_explicit_subdomain",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z06486383UC8WYSBZTWFM"),
							Name: awssdk.String("foo-2.com"),
						},
					},
					nil,
				)
				client.On("ListRecordsForZone", "Z06486383UC8WYSBZTWFM").Return([]*route53.ResourceRecordSet{
					{
						Name: awssdk.String("test0"),
						Type: awssdk.String("TXT"),
					},
					{
						Name: awssdk.String("test0"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("test1.foo-2.com"),
						Type: awssdk.String("TXT"),
					},
					{
						Name: awssdk.String("test1.foo-2.com"),
						Type: awssdk.String("A"),
					},
					{
						Name: awssdk.String("_test2.foo-2.com"),
						Type: awssdk.String("TXT"),
					},
					{
						Name: awssdk.String("_test2.foo-2.com"),
						Type: awssdk.String("A"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list zones",
			dirName: "aws_route53_zone_with_no_record",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					awsError)

				alerter.On("SendAlert", resourceaws.AwsRoute53RecordResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRoute53RecordResourceType, resourceaws.AwsRoute53ZoneResourceType), alerts.EnumerationPhase)).Return()
			},
			err: nil,
		},
		{
			test:    "cannot list records",
			dirName: "aws_route53_zone_with_no_record",
			mocks: func(client *repository.MockRoute53Repository, alerter *mocks.AlerterInterface) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{
						{
							Id:   awssdk.String("Z06486383UC8WYSBZTWFM"),
							Name: awssdk.String("foo-2.com"),
						},
					},
					nil)
				awsError := awserr.NewRequestFailure(awserr.New("AccessDeniedException", "", errors.New("")), 403, "")
				client.On("ListRecordsForZone", "Z06486383UC8WYSBZTWFM").Return(
					[]*route53.ResourceRecordSet{}, awsError)

				alerter.On("SendAlert", resourceaws.AwsRoute53RecordResourceType, alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(awsError, resourceaws.AwsRoute53RecordResourceType, resourceaws.AwsRoute53RecordResourceType), alerts.EnumerationPhase)).Return()
			},
			err: nil,
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
			fakeRepo := &repository.MockRoute53Repository{}
			c.mocks(fakeRepo, alerter)

			var repo repository.Route53Repository = fakeRepo
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
				repo = repository.NewRoute53Repository(session, cache.New(0))
			}

			remoteLibrary.AddEnumerator(aws.NewRoute53RecordEnumerator(repo, factory))
			remoteLibrary.AddDetailsFetcher(resourceaws.AwsRoute53RecordResourceType, common.NewGenericDetailsFetcher(resourceaws.AwsRoute53RecordResourceType, provider, deserializer))

			testFilter := &filter.MockFilter{}
			testFilter.On("IsTypeIgnored", mock.Anything).Return(false)

			s := NewScanner(remoteLibrary, alerter, scanOptions, testFilter)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRoute53RecordResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
			alerter.AssertExpectations(tt)
			fakeRepo.AssertExpectations(tt)
		})
	}
}
