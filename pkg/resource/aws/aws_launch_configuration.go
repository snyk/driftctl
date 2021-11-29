package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

const AwsLaunchConfigurationResourceType = "aws_launch_configuration"

func initAwsLaunchConfigurationMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsLaunchConfigurationResourceType, resource.FlagDeepMode)
}
