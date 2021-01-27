package aws

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type SqsQueueSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       sqsiface.SQSAPI
	runner       *terraform.ParallelResourceReader
}

func NewSqsQueueSupplier(provider *TerraformProvider) *SqsQueueSupplier {
	return &SqsQueueSupplier{
		provider,
		awsdeserializer.NewSqsQueueDeserializer(),
		sqs.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s SqsQueueSupplier) Resources() ([]resource.Resource, error) {
	queues, err := listSqsQueues(s.client)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, aws.AwsSqsQueueResourceType)
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

	return s.deserializer.Deserialize(resources)
}

func (s SqsQueueSupplier) readSqsQueue(queue string) (cty.Value, error) {
	var Ty resource.ResourceType = aws.AwsSqsQueueResourceType
	val, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: Ty,
		ID: queue,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"type": Ty,
		}).Error(err)
		return cty.NilVal, err
	}
	return *val, nil
}

func listSqsQueues(client sqsiface.SQSAPI) ([]*string, error) {
	var queues []*string
	input := sqs.ListQueuesInput{}
	err := client.ListQueuesPages(&input,
		func(resp *sqs.ListQueuesOutput, lastPage bool) bool {
			queues = append(queues, resp.QueueUrls...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}
	return queues, nil
}
