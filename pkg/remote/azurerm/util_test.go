package azurerm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAzurermUtil_trimResourceGroupName(t *testing.T) {
	testcases := []struct {
		name       string
		resourceId string
		expected   string
	}{
		{
			name:       "should return resource's group name",
			resourceId: "/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/api-rg-pro/providers/Microsoft.DBforPostgreSQL/servers/postgresql-server-8791542",
			expected:   "api-rg-pro",
		},
		{
			name:       "should return resource's group name",
			resourceId: "/subscriptions/7bfb2c5c-7308-46ed-8ae4-fffa356eb406/resourceGroups/api-rg-pro",
			expected:   "api-rg-pro",
		},
		{
			name:       "should return resource's group name",
			resourceId: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/foobar/providers/Microsoft.Storage/storageAccounts/testeliedriftctl",
			expected:   "foobar",
		},
		{
			name:       "should return resource's group name",
			resourceId: "/subscriptions/00000000-0000-0000-0000-000000000000/test/foobar/providers/Microsoft.Storage/storageAccounts/testeliedriftctl",
			expected:   "",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			got := trimResourceGroupName(tt.resourceId)
			assert.Equal(t, tt.expected, got)
		})
	}
}
