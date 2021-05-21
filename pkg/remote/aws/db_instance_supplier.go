package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type DBInstanceSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.RDSRepository
	runner       *terraform.ParallelResourceReader
}

func NewDBInstanceSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *DBInstanceSupplier {
	return &DBInstanceSupplier{
		provider,
		deserializer,
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

	return s.deserializer.Deserialize(resourceaws.AwsDbInstanceResourceType, results)
}
