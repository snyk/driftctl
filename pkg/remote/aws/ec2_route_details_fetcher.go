package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

type EC2RouteDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewEC2RouteDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *EC2RouteDetailsFetcher {
	return &EC2RouteDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *EC2RouteDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	attributes := map[string]string{
		"route_table_id": *res.Attributes().GetString("route_table_id"),
	}
	if ipv4 := res.Attributes().GetString("destination_cidr_block"); ipv4 != nil && *ipv4 != "" {
		attributes["destination_cidr_block"] = *ipv4
	}
	if ipv6 := res.Attributes().GetString("destination_ipv6_cidr_block"); ipv6 != nil && *ipv6 != "" {
		attributes["destination_ipv6_cidr_block"] = *ipv6
	}
	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty:         aws.AwsRouteResourceType,
		ID:         res.TerraformId(),
		Attributes: attributes,
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsRouteResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}
	return deserializedRes, nil
}
