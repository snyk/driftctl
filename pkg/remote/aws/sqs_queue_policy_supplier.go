package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type SqsQueuePolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.SQSRepository
	runner       *terraform.ParallelResourceReader
}

func NewSqsQueuePolicySupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, client repository.SQSRepository) *SqsQueuePolicySupplier {
	return &SqsQueuePolicySupplier{
		provider,
		deserializer,
		client,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SqsQueuePolicySupplier) SuppliedType() resource.ResourceType {
	return aws.AwsSqsQueuePolicyResourceType
}

func (s *SqsQueuePolicySupplier) Resources() ([]resource.Resource, error) {
	queues, err := s.client.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), aws.AwsSqsQueueResourceType)
	}

	for _, queue := range queues {
		q := *queue
		s.runner.Run(func() (cty.Value, error) {
			return s.readSqsQueuePolicy(q)
		})
	}

	resources, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(s.SuppliedType(), resources)
}

func (s *SqsQueuePolicySupplier) readSqsQueuePolicy(queueURL string) (cty.Value, error) {
	var Ty resource.ResourceType = s.SuppliedType()
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: Ty,
		ID: queueURL,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": Ty,
		}).Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}
