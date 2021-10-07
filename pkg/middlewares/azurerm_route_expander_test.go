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

func TestAzurermRouteExpander_Execute(t *testing.T) {
	tests := []struct {
		name     string
		input    []*resource.Resource
		expected []*resource.Resource
		mock     func(factory *terraform.MockResourceFactory)
	}{
		{
			name: "test with nil route attribute",
			input: []*resource.Resource{
				{
					Id:   "table1",
					Type: azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": nil,
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:   "table1",
					Type: azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": nil,
					},
				},
			},
		},
		{
			name: "test with empty route attributes",
			input: []*resource.Resource{
				{
					Id:   "table1",
					Type: azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "table1",
					Type:  azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test that resource will not be expanded if it already exist",
			input: []*resource.Resource{
				{
					Id:    "table1/routes/exist",
					Type:  azurerm.AzureRouteResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:   "table1",
					Type: azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{
						"route": []interface{}{
							map[string]interface{}{
								"name": "exist",
							},
						},
					},
				},
			},
			expected: []*resource.Resource{
				{
					Id:    "table1/routes/exist",
					Type:  azurerm.AzureRouteResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "table1",
					Type:  azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{},
				},
			},
		},
		{
			name: "test routes are expanded",
			input: []*resource.Resource{
				{
					Id: "fake_resource",
				},
				{
					Id:   "table1",
					Type: azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{
						"name": "table1",
						"route": []interface{}{
							map[string]interface{}{
								"name": "route1",
							},
							map[string]interface{}{
								"name": "route2",
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
					Id:   "table1",
					Type: azurerm.AzureRouteTableResourceType,
					Attrs: &resource.Attributes{
						"name": "table1",
					},
				},
				{
					Id:    "table1/routes/route1",
					Type:  azurerm.AzureRouteResourceType,
					Attrs: &resource.Attributes{},
				},
				{
					Id:    "table1/routes/route2",
					Type:  azurerm.AzureRouteResourceType,
					Attrs: &resource.Attributes{},
				},
			},
			mock: func(factory *terraform.MockResourceFactory) {
				factory.On(
					"CreateAbstractResource",
					azurerm.AzureRouteResourceType,
					"table1/routes/route1",
					map[string]interface{}{
						"name":             "route1",
						"route_table_name": "table1",
					},
				).Times(1).Return(&resource.Resource{
					Id:    "table1/routes/route1",
					Type:  azurerm.AzureRouteResourceType,
					Attrs: &resource.Attributes{},
				}, nil)
				factory.On(
					"CreateAbstractResource",
					azurerm.AzureRouteResourceType,
					"table1/routes/route2",
					map[string]interface{}{
						"name":             "route2",
						"route_table_name": "table1",
					},
				).Times(1).Return(&resource.Resource{
					Id:    "table1/routes/route2",
					Type:  azurerm.AzureRouteResourceType,
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

			m := NewAzurermRouteExpander(factory)
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
