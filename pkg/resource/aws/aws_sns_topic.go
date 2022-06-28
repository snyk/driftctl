package aws

import (
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

func initSnsTopicMetaData(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	resourceSchemaRepository.SetNormalizeFunc(aws.AwsSnsTopicResourceType, func(res *resource.Resource) {
		val := res.Attrs
		val.DeleteIfDefault("sqs_success_feedback_sample_rate")
		val.DeleteIfDefault("lambda_success_feedback_sample_rate")
		val.DeleteIfDefault("http_success_feedback_sample_rate")
		val.DeleteIfDefault("application_success_feedback_sample_rate")
		val.DeleteIfDefault("firehose_failure_feedback_role_arn")
		val.DeleteIfDefault("firehose_success_feedback_role_arn")
		val.SafeDelete([]string{"name_prefix"})
		val.SafeDelete([]string{"owner"})
	})
}
