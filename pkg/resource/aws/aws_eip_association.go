package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsEipAssociationResourceType = "aws_eip_association"

func initAwsEipAssociationMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsEipAssociationResourceType, resource.FlagDeepMode)
}
