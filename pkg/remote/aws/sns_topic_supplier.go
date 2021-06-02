package aws

import (
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type SNSTopicSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.SNSRepository
	runner       *terraform.ParallelResourceReader
}

func NewSNSTopicSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *SNSTopicSupplier {
	return &SNSTopicSupplier{
		provider,
		deserializer,
		repository.NewSNSClient(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SNSTopicSupplier) Resources() ([]resource.Resource, error) {
	topics, err := s.client.ListAllTopics()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsSnsTopicResourceType)
	}

	for _, topic := range topics {
		t := *topic
		s.runner.Run(func() (cty.Value, error) {
			return s.readTopic(t)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(aws.AwsSnsTopicResourceType, retrieve)
}

func (s *SNSTopicSupplier) readTopic(topic sns.Topic) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *topic.TopicArn,
		Ty: aws.AwsSnsTopicResourceType,
		Attributes: map[string]string{
			"topic_arn": *topic.TopicArn,
		},
	})
	if err != nil {
		logrus.Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
