package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type EC2RouteTableAssociationDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewEC2RouteTableAssociationDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *EC2RouteTableAssociationDetailsFetcher {
	return &EC2RouteTableAssociationDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *EC2RouteTableAssociationDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: aws.AwsRouteTableAssociationResourceType,
		ID: res.TerraformId(),
		Attributes: map[string]string{
			"route_table_id": *res.Attributes().GetString("route_table_id"),
		},
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsRouteTableAssociationResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}
	return deserializedRes, nil
}
