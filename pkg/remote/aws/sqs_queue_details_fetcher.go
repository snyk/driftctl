package aws

import (
	"strings"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
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
		ID: res.TerraformId(),
		Ty: aws.AwsSqsQueueResourceType,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NonExistentQueue") {
			logrus.WithFields(logrus.Fields{
				"id":   res.TerraformId(),
				"type": aws.AwsSqsQueueResourceType,
			}).Debugf("Ignoring queue that seems to be already deleted: %+v", err)
			return nil, nil
		}
		logrus.Error(err)
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType(), res.TerraformId())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsSqsQueueResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
