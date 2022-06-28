package aws

import (
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"github.com/snyk/driftctl/enumeration/terraform"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
)

type SQSQueueDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewSQSQueueDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *SQSQueueDetailsFetcher {
	return &SQSQueueDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *SQSQueueDetailsFetcher) ReadDetails(res *resource.Resource) (*resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: res.ResourceId(),
		Ty: aws.AwsSqsQueueResourceType,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NonExistentQueue") {
			logrus.WithFields(logrus.Fields{
				"id":   res.ResourceId(),
				"type": aws.AwsSqsQueueResourceType,
			}).Debugf("Ignoring queue that seems to be already deleted: %+v", err)
			return nil, nil
		}
		logrus.Error(err)
		return nil, remoteerror.NewResourceScanningError(err, res.ResourceType(), res.ResourceId())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsSqsQueueResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
