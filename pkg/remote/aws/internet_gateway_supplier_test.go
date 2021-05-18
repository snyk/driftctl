package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"

	awssdk "github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	mocks2 "github.com/cloudskiff/driftctl/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInternetGatewaySupplier_Resources(t *testing.T) {
	cases := []struct {
		test    string
		dirName string
		mocks   func(repo *repository.MockEC2Repository)
		err     error
	}{
		{
			test:    "no internet gateways",
			dirName: "internet_gateway_empty",
			mocks: func(repo *repository.MockEC2Repository) {
				repo.On("ListAllInternetGateways").Once().Return([]*ec2.InternetGateway{}, nil)
			},
			err: nil,
		},
		{
			test:    "multiple internet gateways",
			dirName: "internet_gateway_multiple",
			mocks: func(repo *repository.MockEC2Repository) {
				repo.On("ListAllInternetGateways").Once().Return([]*ec2.InternetGateway{
					{
						InternetGatewayId: awssdk.String("igw-025e25487b9ee553b"),
					},
					{
						InternetGatewayId: awssdk.String("igw-0aa10b4e92738d8ed"),
					},
				}, nil)
			},
			err: nil,
		},
		{
			test:    "cannot list internet gateways",
			dirName: "internet_gateway_empty",
			mocks: func(repo *repository.MockEC2Repository) {
				repo.On("ListAllInternetGateways").Once().Return(nil, awserr.NewRequestFailure(nil, 403, ""))
			},
			err: remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsInternetGatewayResourceType),
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
			supplierLibrary.AddSupplier(NewInternetGatewaySupplier(provider))
		}

		t.Run(c.test, func(tt *testing.T) {
			fakeEC2 := repository.MockEC2Repository{}
			c.mocks(&fakeEC2)
			provider := mocks2.NewMockedGoldenTFProvider(c.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			internetGatewayDeserializer := awsdeserializer.NewInternetGatewayDeserializer()
			s := &InternetGatewaySupplier{
				provider,
				internetGatewayDeserializer,
				&fakeEC2,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(tt, c.err, err)

			mock.AssertExpectationsForObjects(tt)
			deserializers := []deserializer.CTYDeserializer{internetGatewayDeserializer}
			test.CtyTestDiffMixed(got, c.dirName, provider, deserializers, shouldUpdate, tt)
		})
	}
}
