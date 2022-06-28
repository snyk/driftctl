package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initAwsAppAutoscalingTargetMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetDiscriminantFunc(aws.AwsAppAutoscalingTargetResourceType, func(self, target *resource.Resource) bool {
		return *self.Attributes().GetString("scalable_dimension") == *target.Attributes().GetString("scalable_dimension")
	})
}
