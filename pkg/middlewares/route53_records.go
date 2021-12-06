package middlewares

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// Remote NS and SAO records from remote state if not managed by IAC
type Route53DefaultZoneRecordSanitizer struct{}

func NewRoute53DefaultZoneRecordSanitizer() Route53DefaultZoneRecordSanitizer {
	return Route53DefaultZoneRecordSanitizer{}
}

func (m Route53DefaultZoneRecordSanitizer) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {

	newRemoteResources := make([]*resource.Resource, 0)

	// We iterate on remote resource and adding them to a new slice except for default records
	// added by aws in the zone at creation
	for _, remoteResource := range *remoteResources {
		existInState := false

		// Ignore all resources other than route53 records
		if remoteResource.ResourceType() != aws.AwsRoute53RecordResourceType {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		if !isDefaultRecord(remoteResource) {
			newRemoteResources = append(newRemoteResources, remoteResource)
			continue
		}

		for _, stateResource := range *resourcesFromState {
			if remoteResource.Equal(stateResource) {
				existInState = true
				break
			}
		}

		if existInState {
			newRemoteResources = append(newRemoteResources, remoteResource)
		}

		if !existInState {
			logrus.WithFields(logrus.Fields{
				"id":   remoteResource.ResourceId(),
				"type": remoteResource.ResourceType(),
			}).Debug("Ignoring default unmanaged record")
		}

	}

	*remoteResources = newRemoteResources

	return nil
}

// Return true if the record is considered as default one added by aws
func isDefaultRecord(record *resource.Resource) bool {
	ty, _ := record.Attrs.Get("type")
	return ty == "NS" || ty == "SOA"
}
