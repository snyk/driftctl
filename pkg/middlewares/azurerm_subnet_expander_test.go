package middlewares

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/azurerm"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/r3labs/diff/v2"
)

func TestAzurermSubnetExpander_Execute(t *testing.T) {
	tests := []struct {
		name     string
		input    []*resource.Resource
		expected []*resource.Resource
		mock     func(factory *terraform.MockResourceFactory)
	}{
		{
			name: "test with nil subnet attribute",
			input: []*resource.Resource{
				{
					Id:   "network1",
					Type: azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{
						"subnet": nil,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "network1",
					Type: azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{
						"subnet": nil,
					},
				},
			},
		},
		{
			name: "test with empty subnet attributes",
			input: []*resource.Resource{
				{
					Id:   "network1",
					Type: azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{
						"subnet": []interface{}{},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "network1",
					Type:  azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test that resource will not be expanded if it already exist",
			input: []*resource.Resource{
				{
					Id:    "exist",
					Type:  azurerm.AzureSubnetResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "network1",
					Type: azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{
						"subnet": []interface{}{
							map[string]interface{}{
								"id": "exist",
							},
						},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "exist",
					Type:  azurerm.AzureSubnetResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "network1",
					Type:  azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test subnet are expanded",
			input: []*resource.Resource{
				{
					Id: "fake_resource",
				},
				{
					Id:   "network1",
					Type: azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{
						"subnet": []interface{}{
							map[string]interface{}{
								"id": "subnet1",
							},
							map[string]interface{}{
								"id": "subnet2",
							},
						},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id: "fake_resource",
				},
				{
					Id:    "network1",
					Type:  azurerm.AzureVirtualNetworkResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "subnet1",
					Type:  azurerm.AzureSubnetResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "subnet2",
					Type:  azurerm.AzureSubnetResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On("CreateAbstractResource", azurerm.AzureSubnetResourceType, "subnet1", map[string]interface{}{}).Times(1).Return(&resource.Resource{
					Id:    "subnet1",
					Type:  azurerm.AzureSubnetResourceType,
					Attrs: &resource.Attributes{},
				}, nil)
				factory.On("CreateAbstractResource", azurerm.AzureSubnetResourceType, "subnet2", map[string]interface{}{}).Times(1).Return(&resource.Resource{
					Id:    "subnet2",
					Type:  azurerm.AzureSubnetResourceType,
					Attrs: &resource.Attributes{},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := &terraform.MockResourceFactory{}
			if tt.mock != nil {
				tt.mock(factory)
			}

			m := NewAzurermSubnetExpander(factory)
			err := m.Execute(&[]*resource.Resource{}, &tt.input)
			if err != nil {
				t.Fatal(err)
			}

			changelog, err := diff.Diff(tt.expected, tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}

		})
	}
}
