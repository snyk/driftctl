package aws

import (
	"context"
	"testing"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/stretchr/testify/assert"

	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
)

func TestDBSubnetGroupSupplier_Resources(t *testing.T) {

	tests := []struct {
		test             string
		dirName          string
		subnets          mocks.DescribeSubnetGroupResponse
		subnetsListError error
		err              error
	}{
		{
			test:    "no subnets",
			dirName: "db_subnet_empty",
			subnets: mocks.DescribeSubnetGroupResponse{
				{
					true,
					&rds.DescribeDBSubnetGroupsOutput{},
				},
			},
			err: nil,
		},
		{
			test:    "multiples db subnets",
			dirName: "db_subnet_multiples",
			subnets: mocks.DescribeSubnetGroupResponse{
				{
					false,
					&rds.DescribeDBSubnetGroupsOutput{
						DBSubnetGroups: []*rds.DBSubnetGroup{
							&rds.DBSubnetGroup{
								DBSubnetGroupName: aws.String("foo"),
							},
						},
					},
				},
				{
					true,
					&rds.DescribeDBSubnetGroupsOutput{
						DBSubnetGroups: []*rds.DBSubnetGroup{
							&rds.DBSubnetGroup{
								DBSubnetGroupName: aws.String("bar"),
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			test:             "Cannot list subnet",
			dirName:          "db_subnet_empty",
			subnetsListError: awserr.NewRequestFailure(nil, 403, ""),
			err:              remoteerror.NewResourceEnumerationError(awserr.NewRequestFailure(nil, 403, ""), resourceaws.AwsDbSubnetGroupResourceType),
		},
	}
	for _, tt := range tests {

		shouldUpdate := tt.dirName == *goldenfile.Update

		providerLibrary := terraform.NewProviderLibrary()
		supplierLibrary := resource.NewSupplierLibrary()

		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			providerLibrary.AddProvider(terraform.AWS, provider)
			supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, providerLibrary.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewDBSubnetGroupDeserializer()
			client := mocks.NewMockAWSRDSSubnetGroupClient(tt.subnets)
			if tt.subnetsListError != nil {
				client = mocks.NewMockAWSRDSErrorClient(tt.subnetsListError)
			}
			s := &DBSubnetGroupSupplier{
				provider,
				deserializer,
				client,
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			assert.Equal(t, tt.err, err)

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
