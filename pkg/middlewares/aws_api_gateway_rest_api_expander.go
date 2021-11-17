package middlewares

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/sirupsen/logrus"
)

// Explodes api gateway rest api body attribute to dedicated resources as per Terraform documentation (https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/api_gateway_rest_api)
type AwsApiGatewayRestApiExpander struct {
	resourceFactory resource.ResourceFactory
}

type OpenAPIAwsExtensions struct {
	GatewayResponses map[string]interface{} `json:"x-amazon-apigateway-gateway-responses"`
}

type OpenAPIAwsMethodExtensions struct {
	Integration map[string]interface{} `json:"x-amazon-apigateway-integration"`
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
		if _, ok := err.(*json.SyntaxError); ok {
			err = yaml.Unmarshal([]byte(*body), &docV3)
		}
		if err != nil {
			return err
		}
	}
	// It's an OpenAPI v3 document
	if docV3.OpenAPI != "" {
		return m.handleBodyV3(api.ResourceId(), docV3, results, remoteResources)
	}

	docV2 := &openapi2.T{}
	if err := json.Unmarshal([]byte(*body), &docV2); err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			err = yaml.Unmarshal([]byte(*body), &docV2)
		}
		if err != nil {
			return err
		}
	}
	// It's an OpenAPI v2 document
	if docV2.Swagger != "" {
		return m.handleBodyV2(api.ResourceId(), docV2, results, remoteResources)
	}

	return nil
}

func (m *AwsApiGatewayRestApiExpander) handleBodyV3(apiId string, doc *openapi3.T, results, remoteResources *[]*resource.Resource) error {
	for path, pathItem := range doc.Paths {
		if res := m.createApiGatewayResource(apiId, path, results, remoteResources); res != nil {
			ops := pathItem.Operations()
			for httpMethod, method := range ops {
				m.createApiGatewayMethod(apiId, res.ResourceId(), httpMethod, results)
				for statusCode := range method.Responses {
					m.createApiGatewayMethodResponse(apiId, res.ResourceId(), httpMethod, statusCode, results)
				}
				m.createApiGatewayIntegration(apiId, res.ResourceId(), httpMethod, results)
				if err := m.createMethodExtensionsResources(apiId, res.ResourceId(), httpMethod, method.Extensions, results); err != nil {
					return nil
				}
			}
		}
	}
	if err := m.createExtensionsResources(apiId, doc.Extensions, results); err != nil {
		return nil
	}
	return nil
}

func (m *AwsApiGatewayRestApiExpander) handleBodyV2(apiId string, doc *openapi2.T, results, remoteResources *[]*resource.Resource) error {
	for path, pathItem := range doc.Paths {
		if res := m.createApiGatewayResource(apiId, path, results, remoteResources); res != nil {
			ops := pathItem.Operations()
			for httpMethod, method := range ops {
				m.createApiGatewayMethod(apiId, res.ResourceId(), httpMethod, results)
				for statusCode := range method.Responses {
					m.createApiGatewayMethodResponse(apiId, res.ResourceId(), httpMethod, statusCode, results)
				}
				m.createApiGatewayIntegration(apiId, res.ResourceId(), httpMethod, results)
				if err := m.createMethodExtensionsResources(apiId, res.ResourceId(), httpMethod, method.Extensions, results); err != nil {
					return nil
				}
			}
		}
	}
	if err := m.createExtensionsResources(apiId, doc.Extensions, results); err != nil {
		return nil
	}
	return nil
}

// Create resources based on our OpenAPIAwsExtensions struct
func (m *AwsApiGatewayRestApiExpander) createExtensionsResources(apiId string, extensions map[string]interface{}, results *[]*resource.Resource) error {
	ext, err := decodeExtensions(extensions)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":   apiId,
			"type": aws.AwsApiGatewayRestApiResourceType,
		}).Debug("Failed to decode extensions from the OpenAPI body attribute")
		return err
	}
	for gtwResponse := range ext.GatewayResponses {
		m.createApiGatewayGatewayResponse(apiId, gtwResponse, results)
	}
	return nil
}

// Create resources based on our OpenAPIAwsMethodExtensions struct
func (m *AwsApiGatewayRestApiExpander) createMethodExtensionsResources(apiId, resourceId, httpMethod string, extensions map[string]interface{}, results *[]*resource.Resource) error {
	ext, err := decodeMethodExtensions(extensions)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":   apiId,
			"type": aws.AwsApiGatewayRestApiResourceType,
		}).Debug("Failed to decode method extensions from the OpenAPI body attribute")
		return err
	}
	if responses, exist := ext.Integration["responses"]; exist {
		for _, response := range responses.(map[string]interface{}) {
			if statusCode, ok := response.(map[string]interface{})["statusCode"]; ok {
				if s, isFloat64 := statusCode.(float64); isFloat64 {
					statusCode = strconv.FormatFloat(s, 'f', -1, 64)
				}
				m.createApiGatewayIntegrationResponse(apiId, resourceId, httpMethod, statusCode.(string), results)
			}
		}
	}
	return nil
}

// Create aws_api_gateway_resource resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayResource(apiId, path string, results, remoteResources *[]*resource.Resource) *resource.Resource {
	if res := foundMatchingResource(apiId, path, remoteResources); res != nil {
		newResource := m.resourceFactory.CreateAbstractResource(aws.AwsApiGatewayResourceResourceType, res.ResourceId(), map[string]interface{}{
			"rest_api_id": *res.Attributes().GetString("rest_api_id"),
			"path":        path,
		})
		*results = append(*results, newResource)
		return newResource
	}
	return nil
}

// Create aws_api_gateway_gateway_response resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayGatewayResponse(apiId, gtwResponse string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsApiGatewayGatewayResponseResourceType,
		strings.Join([]string{"aggr", apiId, gtwResponse}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
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

// Create aws_api_gateway_method resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayMethod(apiId, resourceId, httpMethod string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsApiGatewayMethodResourceType,
		strings.Join([]string{"agm", apiId, resourceId, httpMethod}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// Create aws_api_gateway_method_response resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayMethodResponse(apiId, resourceId, httpMethod, statusCode string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsApiGatewayMethodResponseResourceType,
		strings.Join([]string{"agmr", apiId, resourceId, httpMethod, statusCode}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// Decode openapi.Extensions into our custom OpenAPIAwsExtensions struct that follows AWS
// OpenAPI addons.
func decodeExtensions(extensions map[string]interface{}) (*OpenAPIAwsExtensions, error) {
	rawExtensions, err := json.Marshal(extensions)
	if err != nil {
		return nil, err
	}
	decodedExtensions := &OpenAPIAwsExtensions{}
	err = json.Unmarshal(rawExtensions, decodedExtensions)
	if err != nil {
		return nil, err
	}
	return decodedExtensions, nil
}

// Create aws_api_gateway_integration resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayIntegration(apiId, resourceId, httpMethod string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsApiGatewayIntegrationResourceType,
		strings.Join([]string{"agi", apiId, resourceId, httpMethod}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// Create aws_api_gateway_integration resource
func (m *AwsApiGatewayRestApiExpander) createApiGatewayIntegrationResponse(apiId, resourceId, httpMethod, statusCode string, results *[]*resource.Resource) {
	newResource := m.resourceFactory.CreateAbstractResource(
		aws.AwsApiGatewayIntegrationResponseResourceType,
		strings.Join([]string{"agir", apiId, resourceId, httpMethod, statusCode}, "-"),
		map[string]interface{}{},
	)
	*results = append(*results, newResource)
}

// Decode openapi.Method.Extensions into our custom OpenAPIAwsMethodExtensions struct that follows AWS
// OpenAPI addons.
func decodeMethodExtensions(extensions map[string]interface{}) (*OpenAPIAwsMethodExtensions, error) {
	rawExtensions, err := json.Marshal(extensions)
	if err != nil {
		return nil, err
	}
	decodedExtensions := &OpenAPIAwsMethodExtensions{}
	err = json.Unmarshal(rawExtensions, decodedExtensions)
	if err != nil {
		return nil, err
	}
	return decodedExtensions, nil
}
