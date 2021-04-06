package dctlcty

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCtyAttributes_SafeDelete(t *testing.T) {
	tests := []struct {
		name     string
		attr     CtyAttributes
		path     []string
		expected CtyAttributes
	}{
		{
			name: "Delete existing",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path: []string{"test"},
			expected: map[string]interface{}{
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
		},
		{
			name: "Delete existing nested",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path: []string{"nested", "testNested"},
			expected: map[string]interface{}{
				"test": "exists",

				"nested": map[string]interface{}{},
			},
		},
		{
			name: "Delete not existing",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path: []string{"test1"},
			expected: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
		},
		{
			name: "Delete not existing nested",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path: []string{"nested", "testNest"},
			expected: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
		},
		{
			name: "Delete not real nested",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path: []string{"test", "testNest"},
			expected: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.attr.SafeDelete(tt.path)
			assert.Equal(t, tt.expected, tt.attr)
		})
	}
}

func TestCtyAttributes_SafeSet(t *testing.T) {
	tests := []struct {
		name     string
		attr     CtyAttributes
		path     []string
		value    interface{}
		expected CtyAttributes
		error    error
	}{
		{
			name: "set existing",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path:  []string{"test"},
			value: "CHANGED",
			expected: map[string]interface{}{
				"test": "CHANGED",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
		},
		{
			name: "set existing nested",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path:  []string{"nested", "testNested"},
			value: "CHANGED",
			expected: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "CHANGED",
				},
			},
		},
		{
			name: "Set not existing",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path:  []string{"test1"},
			value: "SET",
			expected: map[string]interface{}{
				"test1": "SET",
				"test":  "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
		},
		{
			name: "SET not existing nested",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path:  []string{"nested", "testNest"},
			value: "SET",
			expected: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
					"testNest":   "SET",
				},
			},
		},
		{
			name: "Delete not real nested",
			attr: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			path:  []string{"test", "testNest"},
			value: "NOK",
			expected: map[string]interface{}{
				"test": "exists",
				"nested": map[string]interface{}{
					"testNested": "exists",
				},
			},
			error: errors.New("Path test.testNest cannot be set: test is not a nested struct"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.attr.SafeSet(tt.path, tt.value)
			if tt.error != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tt.error.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.expected, tt.attr)
		})
	}
}
