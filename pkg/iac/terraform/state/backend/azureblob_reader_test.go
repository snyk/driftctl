package backend

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAzureRMReader(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		preTest func(t *testing.T)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "invalid path",
			path: "containerName/",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, "Unable to parse azurerm backend storage path: containerName/. Must be CONTAINER/PATH/TO/OBJECT", err.Error())
				return true
			},
		},
		{
			name: "valid path but missing AZURE_STORAGE_ACCOUNT",
			path: "containerName/valid.tfstate",
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, "AZURE_STORAGE_ACCOUNT should be defined to be able to read state from azure backend", err.Error())
				return true
			},
		},
		{
			name: "valid path but missing AZURE_STORAGE_KEY",
			path: "containerName/valid.tfstate",
			preTest: func(t *testing.T) {
				t.Setenv("AZURE_STORAGE_ACCOUNT", "foobar")
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				assert.Equal(t, "AZURE_STORAGE_KEY should be defined to be able to read state from azure backend", err.Error())
				return true
			},
		},
		// This is not supposed to do any network call during azure client init
		// It this behavior change that logic should be moved in the Read function like we already
		// did for some other backend
		{
			name: "valid",
			path: "containerName/valid.tfstate",
			preTest: func(t *testing.T) {
				t.Setenv("AZURE_STORAGE_ACCOUNT", "foobar")
				t.Setenv("AZURE_STORAGE_KEY", "barfoo")
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preTest != nil {
				tt.preTest(t)
			}
			_, err := NewAzureRMReader(tt.path)
			if !tt.wantErr(t, err, fmt.Sprintf("NewAzureRMReader(%v)", tt.path)) {
				return
			}
		})
	}
}
