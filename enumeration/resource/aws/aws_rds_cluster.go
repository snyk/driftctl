package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsRDSClusterResourceType = "aws_rds_cluster"

func initAwsRDSClusterMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsRDSClusterResourceType, resource.FlagDeepMode)
}
