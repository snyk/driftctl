package aws

import (
	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type DBInstanceSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       rdsiface.RDSAPI
	runner       *terraform.ParallelResourceReader
}

func NewDBInstanceSupplier(runner *pkg.ParallelRunner, client rdsiface.RDSAPI) *DBInstanceSupplier {
	return &DBInstanceSupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewDBInstanceDeserializer(), client, terraform.NewParallelResourceReader(runner)}
}

func listAwsDBInstances(client rdsiface.RDSAPI) ([]*rds.DBInstance, error) {
	var result []*rds.DBInstance
	input := &rds.DescribeDBInstancesInput{}
	err := client.DescribeDBInstancesPages(input, func(res *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		result = append(result, res.DBInstances...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s DBInstanceSupplier) Resources() ([]resource.Resource, error) {

	resourceList, err := listAwsDBInstances(s.client)

	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, res := range resourceList {
		id := *res.DBInstanceIdentifier
		s.runner.Run(func() (cty.Value, error) {
			completeResource, err := s.reader.ReadResource(terraform.ReadResourceArgs{
				Ty: resourceaws.AwsDbInstanceResourceType,
				ID: id,
			})
			if err != nil {
				logrus.Warnf("Error reading %s[%s]: %+v", id, resourceaws.AwsDbInstanceResourceType, err)
				return cty.NilVal, err
			}
			return *completeResource, nil
		})
	}

	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(results)
}
