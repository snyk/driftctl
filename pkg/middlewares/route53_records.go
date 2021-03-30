package middlewares

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"
)

// Remote NS and SAO records from remote state if not managed by IAC
type Route53DefaultZoneRecordSanitizer struct{}

func NewRoute53DefaultZoneRecordSanitizer() Route53DefaultZoneRecordSanitizer {
	return Route53DefaultZoneRecordSanitizer{}
}

func (m Route53DefaultZoneRecordSanitizer) Execute(remoteResources, resourcesFromState *[]resource.Resource) error {

	newRemoteResources := make([]resource.Resource, 0)

	// We iterate on remote resource and adding them to a new slice except for default records
	// added by aws in the zone at creation
	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than route53 records
		if remoteResource.TerraformType() != aws.AwsRoute53RecordResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		record, _ := remoteResource.(*aws.AwsRoute53Record)

		if !isDefaultRoute53Record(record) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		for _, stateResource := range *resourcesFromState {
			if resource.IsSameResource(remoteResource, stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
		}

		if !existInState {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.TerraformId(),
				"type": remoteResource.TerraformType(),
			}).Debug("Ignoring default unmanaged record")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

// Return true if the record is considered as default one added by aws
func isDefaultRoute53Record(record *aws.AwsRoute53Record) bool {
	return *record.Type == "NS" || *record.Type == "SOA"
}
