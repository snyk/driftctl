package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type DBInstanceSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       repository.RDSRepository
	runner       *terraform.ParallelResourceReader
}

func NewDBInstanceSupplier(provider *AWSTerraformProvider) *DBInstanceSupplier {
	return &DBInstanceSupplier{
		provider,
		awsdeserializer.NewDBInstanceDeserializer(),
		repository.NewRDSRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *DBInstanceSupplier) Resources() ([]resource.Resource, error) {

	resourceList, err := s.client.ListAllDBInstances()

	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsDbInstanceResourceType)
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
