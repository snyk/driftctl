package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testmocks "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

func TestRoute53HealthCheckSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *repository.MockRoute53Repository)
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
					{Id: aws.String("7001a9df-ded4-4802-9909-668eb80b972b")},
					{Id: aws.String("84fc318a-2e0d-41d6-b638-280e2f0f4e26")},
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
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
		resourceaws.InitResourcesMetadata(repo)
		factory := terraform.NewTerraformResourceFactory(repo)

		deserializer := resource.NewDeserializer(factory)
		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewRoute53HealthCheckSupplier(provider, deserializer, repository.NewRoute53Repository(provider.session, cache.New(0))))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeClient := repository.MockRoute53Repository{}
			c.mocks(&fakeClient)
			provider := testmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			s := &Route53HealthCheckSupplier{
				provider,
				deserializer,
				&fakeClient,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)
			mock.AssertExpectationsForObjects(tt)
			test.CtyTestDiff(got, c.dirName, provider, deserializer, shouldUpdate, tt)
		})
	}
}
