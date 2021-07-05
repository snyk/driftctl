package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type Route53ZoneSupplier struct {
	client  repository.Route53Repository
	factory resource.ResourceFactory
}

func NewRoute53ZoneEnumerator(repo repository.Route53Repository, factory resource.ResourceFactory) *Route53ZoneSupplier {
	return &Route53ZoneSupplier{
		repo,
		factory,
	}
}

func (e *Route53ZoneSupplier) SupportedType() resource.ResourceType {
	return resourceaws.AwsRoute53ZoneResourceType
}

func (e *Route53ZoneSupplier) Enumerate() ([]resource.Resource, error) {
	zones, err := e.client.ListAllZones()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsRoute53ZoneResourceType)
	}

	results := make([]resource.Resource, len(zones))

	for _, hostedZone := range zones {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				cleanZoneID(*hostedZone.Id),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}

func cleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}
