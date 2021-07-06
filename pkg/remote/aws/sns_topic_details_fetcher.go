package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type SNSTopicDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewSNSTopicDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *SNSTopicDetailsFetcher {
	return &SNSTopicDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *SNSTopicDetailsFetcher) ReadDetails(topic resource.Resource) (resource.Resource, error) {
	val, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: topic.TerraformId(),
		Ty: aws.AwsSnsTopicResourceType,
		Attributes: map[string]string{
			"topic_arn": topic.TerraformId(),
		},
	})
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	return r.deserializer.DeserializeOne(aws.AwsSnsTopicResourceType, *val)
}
