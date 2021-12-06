package aws

import (
	"github.com/snyk/driftctl/pkg/resource"
)

const AwsSqsQueueResourceType = "aws_sqs_queue"

func initSqsQueueMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetFlags(AwsSqsQueueResourceType, resource.FlagDeepMode)
}
