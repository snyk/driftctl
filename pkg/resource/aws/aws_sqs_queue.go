package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
)

const AwsSqsQueueResourceType = "aws_sqs_queue"

func initSqsQueueMetaData(resourceSchemaRepository dctlresource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsSqsQueueResourceType, resource.FlagDeepMode)
}
