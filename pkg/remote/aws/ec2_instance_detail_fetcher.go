package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
)

type EC2InstanceDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewEC2InstanceDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *EC2InstanceDetailsFetcher {
	return &EC2InstanceDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *EC2InstanceDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsInstanceResourceType,
		ID: res.TerraformId(),
	})
	if err != nil {
		return nil, err
	}
	if ctyVal.IsNull() {
		logrus.WithFields(logrus.Fields{
			"id": res.TerraformId(),
		}).Debug("Instance read returned nil (instance may be terminated), ignoring ...")
		return nil, nil
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsInstanceResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}
	return deserializedRes, nil
}
