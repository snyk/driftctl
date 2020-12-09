package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/route53/route53iface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type Route53ZoneSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       route53iface.Route53API
	runner       *terraform.ParallelResourceReader
}

func NewRoute53ZoneSupplier(runner *pkg.ParallelRunner, client route53iface.Route53API) *Route53ZoneSupplier {
	return &Route53ZoneSupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewRoute53ZoneDeserializer(), client, terraform.NewParallelResourceReader(runner)}
}

func listAwsRoute53Zones(client route53iface.Route53API) ([]*route53.HostedZone, error) {
	var result []*route53.HostedZone
	input := &route53.ListHostedZonesInput{}
	err := client.ListHostedZonesPages(input, func(res *route53.ListHostedZonesOutput, lastPage bool) bool {
		result = append(result, res.HostedZones...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s Route53ZoneSupplier) Resources() ([]resource.Resource, error) {

	zones, err := listAwsRoute53Zones(s.client)
	if err != nil {
		logrus.Error(err)
		return nil, err
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

	return s.deserializer.Deserialize(results)
}

func (s Route53ZoneSupplier) readZone(hostedZone route53.HostedZone) (cty.Value, error) {
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
