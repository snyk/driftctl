package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsRDSClusterResourceType = "aws_rds_cluster"

func initAwsRDSClusterMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsRDSClusterResourceType, func(res *resource.Resource) map[string]string {
		return map[string]string{
			"cluster_identifier": *res.Attributes().GetString("cluster_identifier"),
		}
	})
	resourceSchemaRepository.SetNormalizeFunc(AwsRDSClusterResourceType, func(res *resource.Resource) {
		val := res.Attributes()
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"master_password"})
		val.SafeDelete([]string{"cluster_members"})
		if v := val.GetBool("skip_final_snapshot"); v == nil {
			_ = val.SafeSet([]string{"skip_final_snapshot"}, false)
		}
	})
}
