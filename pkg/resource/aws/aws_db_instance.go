package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsDbInstanceResourceType = "aws_db_instance"

func initAwsDbInstanceMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsDbInstanceResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.SafeDelete([]string{"delete_automated_backups"})
		val.SafeDelete([]string{"final_snapshot_identifier"})
		val.SafeDelete([]string{"latest_restorable_time"})
		val.SafeDelete([]string{"password"})
		val.SafeDelete([]string{"skip_final_snapshot"})
		val.SafeDelete([]string{"timeouts"})
		val.SafeDelete([]string{"snapshot_identifier"})
		val.SafeDelete([]string{"allow_major_version_upgrade"})
		val.SafeDelete([]string{"apply_immediately"})
		val.DeleteIfDefault("CharacterSetName")
	})
}
