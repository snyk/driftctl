package remote

import (
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestRoute53_HealthCheck(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockRoute53Repository)
		err     error
	}{
		{
			test:    "no health check",
			dirName: "route53_health_check_empty",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllHealthChecks").Return([]*route53.HealthCheck{}, nil)
			},
			err: nil,
		},
		{
			test:    "Multiple health check",
			dirName: "route53_health_check_multiple",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllHealthChecks").Return([]*route53.HealthCheck{
					{Id: awssdk.String("7001a9df-ded4-4802-9909-668eb80b972b")},
					{Id: awssdk.String("84fc318a-2e0d-41d6-b638-280e2f0f4e26")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list health check",
			dirName: "route53_health_check_empty",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllHealthChecks").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
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
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockRoute53Repository{}
			c.mocks(fakeRepo)
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

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRoute53HealthCheckResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}

func TestRoute53_Zone(t *testing.T) {

	tests := []struct {
		test    string
		dirName string
		mocks   func(*repository.MockRoute53Repository)
		err     error
	}{
		{
			test:    "no zones",
			dirName: "route53_zone_empty",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					nil,
				)
			},
			err: nil,
		},
		{
			test:    "single zone",
			dirName: "route53_zone_single",
			mocks: func(client *repository.MockRoute53Repository) {
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
			dirName: "route53_zone_multiples",
			mocks: func(client *repository.MockRoute53Repository) {
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
			dirName: "route53_zone_empty",
			mocks: func(client *repository.MockRoute53Repository) {
				client.On("ListAllZones").Return(
					[]*route53.HostedZone{},
					awserr.NewRequestFailure(nil, 403, ""),
				)
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
			alerter.On("SendAlert", mock.Anything, mock.Anything).Maybe().Return()
			fakeRepo := &repository.MockRoute53Repository{}
			c.mocks(fakeRepo)
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

			s := NewScanner(nil, remoteLibrary, alerter, scanOptions)
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			if err != nil {
				return
			}
			test.TestAgainstGoldenFile(got, resourceaws.AwsRoute53ZoneResourceType, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
