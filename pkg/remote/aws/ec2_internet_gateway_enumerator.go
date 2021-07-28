package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

type EC2InternetGatewayEnumerator struct {
	repository repository.EC2Repository
	factory    resource.ResourceFactory
}

func NewEC2InternetGatewayEnumerator(repo repository.EC2Repository, factory resource.ResourceFactory) *EC2InternetGatewayEnumerator {
	return &EC2InternetGatewayEnumerator{
		repository: repo,
		factory:    factory,
	}
}

func (e *EC2InternetGatewayEnumerator) SupportedType() resource.ResourceType {
	return aws.AwsInternetGatewayResourceType
}

func (e *EC2InternetGatewayEnumerator) Enumerate() ([]resource.Resource, error) {
	internetGateways, err := e.repository.ListAllInternetGateways()
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, string(e.SupportedType()))
	}

	results := make([]resource.Resource, len(internetGateways))

	for _, internetGateway := range internetGateways {
		data := map[string]interface{}{}
		if len(internetGateway.Attachments) > 0 && internetGateway.Attachments[0].VpcId != nil {
			data["vpc_id"] = *internetGateway.Attachments[0].VpcId
		}
		results = append(
			results,
			e.factory.CreateAbstractResource(
				string(e.SupportedType()),
				*internetGateway.InternetGatewayId,
				data,
			),
		)
	}

	return results, err
}
