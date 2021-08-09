package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type EC2DefaultRouteTableDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewEC2DefaultRouteTableDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *EC2DefaultRouteTableDetailsFetcher {
	return &EC2DefaultRouteTableDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *EC2DefaultRouteTableDetailsFetcher) ReadDetails(res *resource.Resource) (*resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsDefaultRouteTableResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"vpc_id": *res.Attributes().GetString("vpc_id"),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType(), res.TerraformId())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsDefaultRouteTableResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
