package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type IamAccessKeyDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewIamAccessKeyDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *IamAccessKeyDetailsFetcher {
	return &IamAccessKeyDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *IamAccessKeyDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsIamAccessKeyResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"user": *res.Attributes().GetString("user"),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsIamAccessKeyResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
