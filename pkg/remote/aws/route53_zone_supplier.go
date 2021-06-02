package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type Route53ZoneSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.Route53Repository
	runner       *terraform.ParallelResourceReader
}

func NewRoute53ZoneSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer) *Route53ZoneSupplier {
	return &Route53ZoneSupplier{
		provider,
		deserializer,
		repository.NewRoute53Repository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *Route53ZoneSupplier) Resources() ([]resource.Resource, error) {

	zones, err := s.client.ListAllZones()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsRoute53ZoneResourceType)
	}

	for _, hostedZone := range zones {
		zone := *hostedZone
		s.runner.Run(func() (cty.Value, error) {
			return s.readZone(zone)
		})
	}
	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(resourceaws.AwsRoute53ZoneResourceType, results)
}

func (s *Route53ZoneSupplier) readZone(hostedZone route53.HostedZone) (cty.Value, error) {
	name := *hostedZone.Name

	zone, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsRoute53ZoneResourceType,
		ID: cleanZoneID(*hostedZone.Id),
	})
	if err != nil {
		logrus.Warnf("Error reading %s[%s]: %+v", name, resourceaws.AwsRoute53ZoneResourceType, err)
		return cty.NilVal, err
	}

	return *zone, nil
}

func cleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}
