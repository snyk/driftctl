package aws

import (
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	remoteerror "github.com/snyk/driftctl/enumeration/remote/error"
	"strings"

	"github.com/snyk/driftctl/enumeration/resource"
	resourceaws "github.com/snyk/driftctl/enumeration/resource/aws"
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

func (e *Route53ZoneSupplier) Enumerate() ([]*resource.Resource, error) {
	zones, err := e.client.ListAllZones()
	if err != nil {
		return nil, remoteerror.NewResourceListingError(err, string(e.SupportedType()))
	}

	results := make([]*resource.Resource, 0, len(zones))

	for _, hostedZone := range zones {
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				strings.TrimPrefix(*hostedZone.Id, "/hostedzone/"),
				map[string]interface{}{},
			),
		)
	}

	return results, err
}
