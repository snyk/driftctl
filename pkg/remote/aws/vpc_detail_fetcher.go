package aws

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type VPCDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewVPCDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *VPCDetailsFetcher {
	return &VPCDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *VPCDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resource.ResourceType(res.TerraformType()),
		ID: res.TerraformId(),
	})
	if err != nil {
		return nil, err
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsVpcResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
