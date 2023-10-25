package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsRDSClusterResourceType = "aws_rds_cluster"

func initAwsRDSClusterMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsRDSClusterResourceType, func(res *resource.Resource) {
		val := res.Attributes()
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"master_password"})
		val.SafeDelete([]string{"cluster_members"})
		val.SafeDelete([]string{"skip_final_snapshot"})
		val.SafeDelete([]string{"allow_major_version_upgrade"})
		val.SafeDelete([]string{"apply_immediately"})
		val.SafeDelete([]string{"final_snapshot_identifier"})
		val.SafeDelete([]string{"source_region"})
	})
}
