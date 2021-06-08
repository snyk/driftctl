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

func NewSNSTopicSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, client repository.SNSRepository) *SNSTopicSupplier {
	return &SNSTopicSupplier{
		provider,
		deserializer,
		client,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SNSTopicSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsSnsTopicResourceType
}

func (s *SNSTopicSupplier) Resources() ([]resource.Resource, error) {
	topics, err := s.client.ListAllTopics()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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

	return s.deserializer.Deserialize(s.SuppliedType(), retrieve)
}

func (s *SNSTopicSupplier) readTopic(topic sns.Topic) (cty.Value, error) {
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		ID: *topic.TopicArn,
		Ty: s.SuppliedType(),
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
