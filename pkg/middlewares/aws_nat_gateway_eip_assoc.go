package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

type AwsNatGatewayEipAssoc struct{}

func NewAwsNatGatewayEipAssoc() AwsNatGatewayEipAssoc {
	return AwsNatGatewayEipAssoc{}
}

// When creating a nat gateway, we associate an EIP to the gateway
// It implies that driftctl read a aws_eip_association resource from remote
// As we cannot use aws_eip_association in terraform to assign an eip to an aws_nat_gateway
// we should remove this association to ensure we do not output noise in unmanaged resources
func (a AwsNatGatewayEipAssoc) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0, len(*remoteResources))

	for _, remoteResource := range *remoteResources {
		// Ignore all resources other than aws_eip_association
		if remoteResource.ResourceType() != aws.AwsEipAssociationResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		isAssociatedToNatGateway := false

		// Search for a nat gateway associated with our EIP
		for _, res := range *remoteResources {
			if res.ResourceType() == aws.AwsNatGatewayResourceType {
				allocationId, allocationIdExist := res.Attrs.Get("allocation_id")
				eipAssocAllocId, eipAssocAllocIdExist := remoteResource.Attrs.Get("allocation_id")
				if allocationIdExist && eipAssocAllocIdExist &&
					allocationId == eipAssocAllocId {
					isAssociatedToNatGateway = true
					break
				}
			}
		}

		if isAssociatedToNatGateway {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.ResourceId(),
				"type": remoteResource.ResourceType(),
			}).Debug("Ignoring aws_eip_association as it is associated to a nat gateway")
			continue
		}

		newRemoteResources = append(newRemoteResources, remoteResource)
	}

	*remoteResources = newRemoteResources

	return nil
}
