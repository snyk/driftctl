package middlewares

import (
	"encoding/json"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
)

// Explodes api gateway rest api body attribute to dedicated resources as per Terraform documentation (https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_rest_api)
type AwsApiGatewayRestApiExpander struct {
	resourceFactory resource.ResourceFactory
}

func NewAwsApiGatewayRestApiExpander(resourceFactory resource.ResourceFactory) AwsApiGatewayRestApiExpander {
	return AwsApiGatewayRestApiExpander{
		resourceFactory: resourceFactory,
	}
}

func (m AwsApiGatewayRestApiExpander) Execute(remoteResources, resourcesFromState *[]*resource.Resource) error {
	newStateResources := make([]*resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than aws_api_gateway_rest_api
		if res.ResourceType() != aws.AwsApiGatewayRestApiResourceType {
			newStateResources = append(newStateResources, res)
			continue
		}

		newStateResources = append(newStateResources, res)

		err := m.handleBody(res, &newStateResources, remoteResources)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newStateResources
	return nil
}

func (m *AwsApiGatewayRestApiExpander) handleBody(api *resource.Resource, results, remoteResources *[]*resource.Resource) error {
	body := api.Attrs.GetString("body")
	if body == nil || *body == "" {
		return nil
	}

	docV3 := &openapi3.T{}
	if err := json.Unmarshal([]byte(*body), &docV3); err != nil {
		return err
	}
	// It's an OpenAPI v3 document
	if docV3.OpenAPI != "" {
		return m.handleBodyV3(api.ResourceId(), docV3, results, remoteResources)
	}

	docV2 := &openapi2.T{}
	if err := json.Unmarshal([]byte(*body), &docV2); err != nil {
		return err
	}
	// It's an OpenAPI v2 document
	if docV2.Swagger != "" {
		return m.handleBodyV2(api.ResourceId(), docV2, results, remoteResources)
	}

	return nil
}

func (m *AwsApiGatewayRestApiExpander) handleBodyV3(apiId string, doc *openapi3.T, results, remoteResources *[]*resource.Resource) error {
	for path := range doc.Paths {
		m.createApiGatewayResource(apiId, path, results, remoteResources)
	}
	return nil
}

func (m *AwsApiGatewayRestApiExpander) handleBodyV2(apiId string, doc *openapi2.T, results, remoteResources *[]*resource.Resource) error {
	for path := range doc.Paths {
		m.createApiGatewayResource(apiId, path, results, remoteResources)
	}
	return nil
}

// Create aws_api_gateway_resource resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayResource(apiId, path string, results, remoteResources *[]*resource.Resource) {
	if res := foundMatchingResource(apiId, path, remoteResources); res != nil {
		newResource := m.resourceFactory.CreateAbstractResource(aws.AwsApiGatewayResourceResourceType, res.ResourceId(), map[string]interface{}{
			"rest_api_id": *res.Attributes().GetString("rest_api_id"),
			"path":        path,
		})
		*results = append(*results, newResource)
	}
}

// Returns the aws_api_gateway_resource resource that matches the path attribute
func foundMatchingResource(apiId, path string, remoteResources *[]*resource.Resource) *resource.Resource {
	for _, res := range *remoteResources {
		if res.ResourceType() == aws.AwsApiGatewayResourceResourceType {
			p := res.Attributes().GetString("path")
			i := res.Attributes().GetString("rest_api_id")
			if p != nil && i != nil && *p == path && *i == apiId {
				return res
			}
		}
	}
	return nil
}
