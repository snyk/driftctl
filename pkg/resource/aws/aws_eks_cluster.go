package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
)

const AwsEKSClusterResourceType = "aws_eks_cluster"

func initAwsEKSClusterMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(AwsEKSClusterResourceType, func(res *resource.AbstractResource) {
		val := res.Attrs
		val.SafeDelete([]string{"created_at"})
		val.SafeDelete([]string{"timeouts"})
	})
	resourceSchemaRepository.SetResolveReadAttributesFunc(AwsEKSClusterResourceType, func(res *resource.AbstractResource) map[string]string {
		return map[string]string{
			"id":   res.TerraformId(),
			"name": res.TerraformId(),
		}
	})
}
