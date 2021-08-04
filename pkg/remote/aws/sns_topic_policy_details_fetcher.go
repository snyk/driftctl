package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type SNSTopicPolicyDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewSNSTopicPolicyDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *SNSTopicPolicyDetailsFetcher {
	return &SNSTopicPolicyDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *SNSTopicPolicyDetailsFetcher) ReadDetails(topic resource.Resource) (resource.Resource, error) {
	val, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: topic.TerraformId(),
		Ty: aws.AwsSnsTopicPolicyResourceType,
		Attributes: map[string]string{
			"topic_arn": topic.TerraformId(),
		},
	})
	if err != nil {
		logrus.Error(err)
		return nil, remoteerror.NewResourceScanningError(err, topic.TerraformType(), topic.TerraformId())
	}
	return r.deserializer.DeserializeOne(aws.AwsSnsTopicPolicyResourceType, *val)
}
