package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type SNSTopicSubscriptionDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewSNSTopicSubscriptionDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *SNSTopicSubscriptionDetailsFetcher {
	return &SNSTopicSubscriptionDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *SNSTopicSubscriptionDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: res.TerraformId(),
		Ty: aws.AwsSnsTopicSubscriptionResourceType,
		Attributes: map[string]string{
			"SubscriptionId": res.TerraformId(),
		},
	})
	if err != nil {
		logrus.Error(err)
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsSnsTopicSubscriptionResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
