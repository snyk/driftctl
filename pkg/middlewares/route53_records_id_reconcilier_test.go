package middlewares

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/stretchr/testify/assert"
)

func TestRoute53RecordIDReconcilier_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []*resource.Resource
		expected           []*resource.Resource
	}{
		{
			name: "test that id are normalized",
			resourcesFromState: []*resource.Resource{
				{},
				{
					Id:   "1234_toto_TXT",
					Type: aws.AwsRoute53RecordResourceType,
					Attrs: &resource.Attributes{
						"id":      "1234_toto_TXT",
						"zone_id": "1234",
						"fqdn":    "toto.example.com",
						"type":    "TXT",
					},
				},
			},
			expected: []*resource.Resource{
				{},
				{
					Id:   "1234_toto.example.com_TXT",
					Type: aws.AwsRoute53RecordResourceType,
					Attrs: &resource.Attributes{
						"id":      "1234_toto.example.com_TXT",
						"zone_id": "1234",
						"fqdn":    "toto.example.com",
						"type":    "TXT",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewRoute53RecordIDReconcilier()
			err := m.Execute(nil, &tt.resourcesFromState)

			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tt.expected, tt.resourcesFromState)

		})
	}
}
