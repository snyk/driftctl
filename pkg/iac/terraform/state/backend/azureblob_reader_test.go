package backend

import (
	"fmt"
	"testing"

	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend/options"
	"github.com/stretchr/testify/assert"
)

func TestNewAzureRMReader(t *testing.T) {
	tests := []struct {
		name    string
		options options.AzureRMBackendOptions
		path    string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "invalid path",
			path: "containerName/",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.NotNil(t, err)
				assert.Equal(t, "Unable to parse azurerm backend storage path: containerName/. Must be CONTAINER/PATH/TO/OBJECT", err.Error())
				return true
			},
		},
		// This is not supposed to do any network call during azure client init
		// It this behavior change that logic should be moved in the Read function like we already
		// did for some other backend
		{
			name: "valid",
			path: "containerName/valid.tfstate",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAzureRMReader(tt.path, tt.options)
			if !tt.wantErr(t, err, fmt.Sprintf("NewAzureRMReader(%v)", tt.path)) {
				return
			}
		})
	}
}
