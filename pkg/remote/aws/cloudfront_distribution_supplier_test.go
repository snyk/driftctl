package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/aws/aws-sdk-go/aws/awserr"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	testmocks "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCloudfrontDistributionSupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(client *mocks.CloudfrontRepository)
		err     error
	}{
		{
			test:    "no cloudfront distribution",
			dirName: "cloudfront_distribution_empty",
			mocks: func(client *mocks.CloudfrontRepository) {
				client.On("ListAllDistributions").Return([]*cloudfront.DistributionSummary{}, nil)
			},
			err: nil,
		},
		{
			test:    "one cloudfront distribution",
			dirName: "cloudfront_distribution_one",
			mocks: func(client *mocks.CloudfrontRepository) {
				client.On("ListAllDistributions").Return([]*cloudfront.DistributionSummary{
					{Id: aws.String("E1M9CNS0XSHI19")},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list cloudfront distributions",
			dirName: "cloudfront_distribution_empty",
			mocks: func(client *mocks.CloudfrontRepository) {
				client.On("ListAllDistributions").Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsCloudfrontDistributionResourceType),
		},
	}
	for _, c := range cases {
		shouldUpdate := c.dirName == *goldenfile.Update
		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := InitTestAwsProvider(providerLibrary)
			if err != nil {
				t.Fatal(err)
			}
			supplierLibrary.AddSupplier(NewCloudfrontDistributionSupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeCloudfront := mocks.CloudfrontRepository{}
			c.mocks(&fakeCloudfront)
			provider := testmocks.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			cloudfrontDistributionDeserializer := awsdeserializer.NewCloudfrontDistributionDeserializer()
			s := &CloudfrontDistributionSupplier{
				provider,
				cloudfrontDistributionDeserializer,
				&fakeCloudfront,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			deserializers := []deserializer.CTYDeserializer{cloudfrontDistributionDeserializer}
			test.CtyTestDiffMixed(got, c.dirName, provider, deserializers, shouldUpdate, tt)
		})
	}
}
