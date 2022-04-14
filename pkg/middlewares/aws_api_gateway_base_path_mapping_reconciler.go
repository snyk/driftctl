package middlewares

import (
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
)

// AwsApiGatewayBasePathMappingReconciler is used to reconcile API Gateway base path mapping (v1 and v2)
// from both remote and state resources because v1|v2 AWS SDK list endpoints return all mappings
// without distinction.
type AwsApiGatewayBasePathMappingReconciler struct{}

func NewAwsApiGatewayBasePathMappingReconciler() AwsApiGatewayBasePathMappingReconciler {
	return AwsApiGatewayBasePathMappingReconciler{}
}

func (m AwsApiGatewayBasePathMappingReconciler) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newRemoteResources := make([]*resource.Resource, 0)
	managedApiMapping := make([]*resource.Resource, 0)
	unmanagedApiMapping := make([]*resource.Resource, 0)
	for _, res := range *remoteResources {
		// Ignore all resources other than aws_api_gateway_base_path_mapping and aws_apigatewayv2_api_mapping
		if res.ResourceType() != aws.AwsApiGatewayBasePathMappingResourceType &&
			res.ResourceType() != aws.AwsApiGatewayV2MappingResourceType {
			newRemoteResources = append(newRemoteResources, res)
			continue
		}

		// Find a matching state resources
		existInState := false
		for _, stateResource := range *resourcesFromState {
			if res.Equal(stateResource) {
				existInState = true
				break
			}
		}

		// Keep track of the resource if it's managed in IaC
		if existInState {
			managedApiMapping = append(managedApiMapping, res)
			continue
		}

		// If we're here, it means that we are left with unmanaged path mappings
		// in both v1 and v2 format. Let's group real and duplicate path mappings
		// in a slice
		unmanagedApiMapping = append(unmanagedApiMapping, res)
	}

	// We only want to show to our end users unmanaged v1 path mappings
	// To do that, we're going to loop over unmanaged path mappings to delete duplicates
	// and leave after that only v1 path mappings (e.g. remove v2 ones)
	deduplicatedUnmanagedMappings := make([]*resource.Resource, 0, len(unmanagedApiMapping))
	for _, unmanaged := range unmanagedApiMapping {
		// Remove duplicates (e.g. same id, the opposite type)
		isDuplicate := false
		for _, managed := range managedApiMapping {
			if managed.ResourceId() == unmanaged.ResourceId() {
				isDuplicate = true
				break
			}
		}
		if isDuplicate {
			continue
		}

		// Now keep only v1 path mappings
		if unmanaged.ResourceType() == aws.AwsApiGatewayBasePathMappingResourceType {
			deduplicatedUnmanagedMappings = append(deduplicatedUnmanagedMappings, unmanaged)
		}
	}

	// Finally, add both managed and unmanaged resources to remote resources
	newRemoteResources = append(newRemoteResources, managedApiMapping...)
	newRemoteResources = append(newRemoteResources, deduplicatedUnmanagedMappings...)

	*remoteResources = newRemoteResources
	return nil
}
