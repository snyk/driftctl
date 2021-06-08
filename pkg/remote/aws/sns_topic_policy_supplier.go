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

type SNSTopicPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.SNSRepository
	runner       *terraform.ParallelResourceReader
}

func NewSNSTopicPolicySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.SNSRepository) *SNSTopicPolicySupplier {
	return &SNSTopicPolicySupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SNSTopicPolicySupplier) SuppliedType() resource.ResourceType {
	return aws.AwsSnsTopicPolicyResourceType
}

func (s *SNSTopicPolicySupplier) Resources() ([]resource.Resource, error) {
	topics, err := s.client.ListAllTopics()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), aws.AwsSnsTopicResourceType)
	}

	for _, topic := range topics {
		topic := *topic
		s.runner.Run(func() (cty.Value, error) {
			return s.readTopicPolicy(topic)
		})
	}

	retrieve, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(s.SuppliedType(), retrieve)
}

func (s *SNSTopicPolicySupplier) readTopicPolicy(topic sns.Topic) (cty.Value, error) {
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
