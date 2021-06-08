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

type SqsQueueSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.SQSRepository
	runner       *terraform.ParallelResourceReader
}

func NewSqsQueueSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, client repository.SQSRepository) *SqsQueueSupplier {
	return &SqsQueueSupplier{
		provider,
		deserializer,
		client,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *SqsQueueSupplier) SuppliedType() resource.ResourceType {
	return aws.AwsSqsQueueResourceType
}

func (s *SqsQueueSupplier) Resources() ([]resource.Resource, error) {
	queues, err := s.client.ListAllQueues()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}

	for _, queue := range queues {
		q := *queue
		s.runner.Run(func() (cty.Value, error) {
			return s.readSqsQueue(q)
		})
	}

	resources, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(s.SuppliedType(), resources)
}

func (s *SqsQueueSupplier) readSqsQueue(queueURL string) (cty.Value, error) {
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
