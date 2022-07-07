package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

const AwsRDSClusterResourceType = "aws_rds_cluster"

func initAwsRDSClusterMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsRDSClusterResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"cluster_identifier": *res.Attributes().GetString("cluster_identifier"),
			"database_name":      *res.Attributes().GetString("database_name"),
		}
	})
	resourceSchemaRepository.SetFlags(AwsRDSClusterResourceType, resource.FlagDeepMode)
}
