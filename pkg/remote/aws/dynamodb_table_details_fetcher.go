package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type DynamoDBTableDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewDynamoDBTableDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *DynamoDBTableDetailsFetcher {
	return &DynamoDBTableDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *DynamoDBTableDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resource.ResourceType(res.TerraformType()),
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"table_name": res.TerraformId(),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsDynamodbTableResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
