package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type EC2EipAssociationSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       ec2iface.EC2API
	runner       *terraform.ParallelResourceReader
}

func NewEC2EipAssociationSupplier(provider *TerraformProvider) *EC2EipAssociationSupplier {
	return &EC2EipAssociationSupplier{
		provider,
		awsdeserializer.NewEC2EipAssociationDeserializer(),
		ec2.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner())}
}

func (s EC2EipAssociationSupplier) Resources() ([]resource.Resource, error) {
	associationIds, err := listAddressesAssociationIds(s.client)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsEipAssociationResourceType)
	}
	results := make([]cty.Value, 0)
	if len(associationIds) > 0 {
		for _, assocId := range associationIds {
			s.runner.Run(func() (cty.Value, error) {
				return s.readEIPAssociation(assocId)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s EC2EipAssociationSupplier) readEIPAssociation(assocId string) (cty.Value, error) {
	resAssoc, err := s.reader.ReadResource(terraform.ReadResourceArgs{
		Ty: resourceaws.AwsEipAssociationResourceType,
		ID: assocId,
	})
	if err != nil {
		logrus.Warnf("Error reading eip association %s[%s]: %+v", assocId, resourceaws.AwsEipAssociationResourceType, err)
		return cty.NilVal, err
	}
	return *resAssoc, nil
}

func listAddressesAssociationIds(client ec2iface.EC2API) ([]string, error) {
	results := make([]string, 0)
	addresses, err := listAddresses(client)
	if err != nil {
		return nil, err
	}
	for _, address := range addresses {
		if address.AssociationId != nil {
			results = append(results, aws.StringValue(address.AssociationId))
		}
	}
	return results, nil
}
