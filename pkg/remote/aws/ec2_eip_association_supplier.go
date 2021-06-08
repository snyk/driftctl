package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EipAssociationSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	client       repository.EC2Repository
	runner       *terraform.ParallelResourceReader
}

func NewEC2EipAssociationSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.EC2Repository) *EC2EipAssociationSupplier {
	return &EC2EipAssociationSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner())}
}

func (s *EC2EipAssociationSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsEipAssociationResourceType
}

func (s *EC2EipAssociationSupplier) Resources() ([]resource.Resource, error) {
	associationIds, err := s.client.ListAllAddressesAssociation()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
	}
	results := make([]cty.Value, 0)
	if len(associationIds) > 0 {
		for _, assocId := range associationIds {
			assocId := assocId
			s.runner.Run(func() (cty.Value, error) {
				return s.readEIPAssociation(assocId)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *EC2EipAssociationSupplier) readEIPAssociation(assocId string) (cty.Value, error) {
	resAssoc, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: s.SuppliedType(),
		ID: assocId,
	})
	if err != nil {
		logrus.Warnf("Error reading eip association %s[%s]: %+v", assocId, s.SuppliedType(), err)
		return cty.NilVal, err
	}
	return *resAssoc, nil
}
