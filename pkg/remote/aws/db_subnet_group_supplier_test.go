package aws

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/parallel"
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
		test    string
		dirName string
		subnets mocks.DescribeSubnetGroupResponse
		err     error
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
	}
	for _, tt := range tests {
		shouldUpdate := tt.dirName == *goldenfile.Update
		if shouldUpdate {
			provider, err := NewTerraFormProvider()
			if err != nil {
				t.Fatal(err)
			}

			terraform.AddProvider(terraform.AWS, provider)
			resource.AddSupplier(NewDBInstanceSupplier(provider.Runner(), rds.New(provider.session)))
		}

		t.Run(tt.test, func(t *testing.T) {
			provider := mocks.NewMockedGoldenTFProvider(tt.dirName, terraform.Provider(terraform.AWS), shouldUpdate)
			deserializer := awsdeserializer.NewDBSubnetGroupDeserializer()
			s := &DBSubnetGroupSupplier{
				provider,
				deserializer,
				mocks.NewMockAWSRDSSubnetGroupClient(tt.subnets),
				terraform.NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10)),
			}
			got, err := s.Resources()
			if tt.err != err {
				t.Errorf("Expected error %+v got %+v", tt.err, err)
			}

			test.CtyTestDiff(got, tt.dirName, provider, deserializer, shouldUpdate, t)
		})
	}
}
