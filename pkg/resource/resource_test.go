package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Normalize empty slices and map to nil
func TestSanitizeDefaults(t *testing.T) {
	cases := map[string]struct {
		input    Attributes
		expected interface{}
	}{
		"simple": {
			input: Attributes{
				"emptyStringSlice": []string{},
				"emptyIntSlice":    []int{},
				"emptyBoolSlice":   []bool{},
				"emptyMap":         map[string]string{},
				"nilInterface":     interface{}(nil),
				"not_deleted":      "value",
			},
			expected: Attributes{
				"emptyStringSlice": nil,
				"emptyIntSlice":    nil,
				"emptyBoolSlice":   nil,
				"emptyMap":         nil,
				"nilInterface":     nil,
				"not_deleted":      "value",
			},
		},
		"nested": {
			input: Attributes{
				"should": map[string]interface{}{
					"be_deleted":     map[string]interface{}{},
					"be_deleted_too": []string{},
					"not_be_deleted": "no",
				},
				"not_deleted": "value",
			},
			expected: Attributes{
				"should": map[string]interface{}{
					"be_deleted":     nil,
					"be_deleted_too": nil,
					"not_be_deleted": "no",
				},
				"not_deleted": "value",
			},
		},
		"nested_slice": {
			input: Attributes{
				"should": []map[string][]interface{}{
					{
						"be": []interface{}{
							map[string]interface{}{
								"removed":     []string{},
								"removed_too": map[string]string{},
							},
						},
					},
				},
			},
			expected: Attributes{
				"should": []map[string][]interface{}{
					{
						"be": []interface{}{
							map[string]interface{}{
								"removed":     nil,
								"removed_too": nil,
							},
						},
					},
				},
			},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			c.input.SanitizeDefaults()
			assert.Equal(t, c.expected, c.input)
		})
	}
}

// Delete empty or nil slices and maps
func TestSanitizeDefaultsV3(t *testing.T) {
	cases := map[string]struct {
		input    Attributes
		expected interface{}
	}{
		"simple": {
			input: Attributes{
				"emptyStringSlice": []string{},
				"emptyIntSlice":    []int{},
				"emptyBoolSlice":   []bool{},
				"emptyMap":         map[string]string{},
				"nilInterface":     interface{}(nil),
				"not_deleted":      "value",
			},
			expected: Attributes{
				"not_deleted": "value",
			},
		},
		"nested": {
			input: Attributes{
				"should": map[string]interface{}{
					"be_deleted":         map[string]interface{}{},
					"be_deleted_too":     []string{},
					"not_be_deleted":     "no",
					"not_be_deleted_too": []string(nil),
				},
				"not_deleted": "value",
			},
			expected: Attributes{
				"should": map[string]interface{}{
					"not_be_deleted": "no",
				},
				"not_deleted": "value",
			},
		},
		"nested_slice": {
			input: Attributes{
				"should": []map[string][]interface{}{
					{
						"be": []interface{}{
							map[string]interface{}{
								"removed":     []string{},
								"removed_too": map[string]string{},
							},
						},
					},
				},
			},
			expected: Attributes{
				"should": []map[string][]interface{}{
					{
						"be": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
		},
	}
	for k, c := range cases {
		t.Run(k, func(t *testing.T) {
			c.input.SanitizeDefaultsV3()
			assert.Equal(t, c.expected, c.input)
		})
	}
}
