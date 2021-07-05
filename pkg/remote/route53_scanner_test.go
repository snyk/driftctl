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
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	terraform2 "github.com/cloudskiff/driftctl/test/terraform"

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
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsRoute53HealthCheckResourceType),
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

			session := session.Must(session.NewSessionWithOptions(session.Options{
				SharedConfigState: session.SharedConfigEnable,
			}))

			scanOptions := ScannerOptions{Deep: true}
			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()

			// Initialize mocks
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
