package aws

import (
	"strings"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/zclconf/go-cty/cty"
)

type Route53RecordSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.Route53Repository
	runner       *terraform.ParallelResourceReader
}

func NewRoute53RecordSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.Route53Repository) *Route53RecordSupplier {
	return &Route53RecordSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner())}
}

func (s *Route53RecordSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsRoute53RecordResourceType
}

func (s *Route53RecordSupplier) Resources() ([]resource.Resource, error) {

	zones, err := s.listZones()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), resourceaws.AwsRoute53ZoneResourceType)
	}

	for _, zone := range zones {
		if err := s.listRecordsForZone(zone[0], zone[1]); err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
		}
	}

	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *Route53RecordSupplier) listZones() ([][2]string, error) {
	results := make([][2]string, 0)
	zones, err := s.client.ListAllZones()
	if err != nil {
		return nil, err
	}

	for _, hostedZone := range zones {
		results = append(results, [2]string{strings.TrimPrefix(*hostedZone.Id, "/hostedzone/"), *hostedZone.Name})
	}

	return results, nil
}

func (s *Route53RecordSupplier) listRecordsForZone(zoneId string, zoneName string) error {

	records, err := s.client.ListRecordsForZone(zoneId)

	if err != nil {
		return err
	}

	for _, raw := range records {
		rawType := *raw.Type
		rawName := *raw.Name
		rawSetIdentifier := raw.SetIdentifier
		s.runner.Run(func() (cty.Value, error) {
			vars := []string{
				zoneId,
				strings.ToLower(strings.TrimSuffix(rawName, ".")),
				rawType,
			}
			if rawSetIdentifier != nil {
				vars = append(vars, *rawSetIdentifier)
			}

			record, err := s.reader.ReadResource(
				terraform.ReadResourceArgs{
					Ty: s.SuppliedType(),
					ID: strings.Join(vars, "_"),
				},
			)
			if err != nil {
				return cty.NilVal, err
			}

			return *record, nil
		})

	}
	return nil
}
