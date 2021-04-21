package dctlcty

import (
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/configs/configschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/zclconf/go-cty/cty/gocty"
)

func TestCtyAttributes_SafeDelete(t *testing.T) {
	tests := []struct {
		name     string
		attr     map[string]interface{}
		path     []string
		expected map[string]interface{}
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
			attr := CtyAttributes{
				Attrs: tt.attr,
			}
			attr.SafeDelete(tt.path)
			assert.Equal(t, tt.expected, tt.attr)
		})
	}
}

func TestCtyAttributes_SafeSet(t *testing.T) {
	tests := []struct {
		name     string
		attr     map[string]interface{}
		path     []string
		value    interface{}
		expected map[string]interface{}
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
			attr := CtyAttributes{
				Attrs: tt.attr,
			}
			err := attr.SafeSet(tt.path, tt.value)
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

type args struct {
	Attrs    map[string]interface{}
	metadata *ResourceMetadata
}

func TestCtyAttributes_Tags(t *testing.T) {
	tests := []struct {
		name string
		args *args
		path []string
		want *AttributeMetadata
	}{
		{
			"Found tags",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {Configshema: configschema.Attribute{
							Computed: true,
						}},
					},
				},
			},
			[]string{"test", "has", "tags"},
			&AttributeMetadata{Configshema: configschema.Attribute{
				Computed: true,
			}},
		},
		{
			"No tags found",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {Configshema: configschema.Attribute{
							Computed: true,
						}},
					},
				},
			},
			[]string{"test", "has", "no", "tags"},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := gocty.ImpliedType(tt.args.Attrs)
			ctyVal, _ := gocty.ToCtyValue(tt.args.Attrs, typ)

			attrs := CtyAttributes{
				Attrs:    tt.args.Attrs,
				value:    &ctyVal,
				metadata: tt.args.metadata,
			}
			if got := attrs.AttributeMetadata(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AttributeMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCtyAttributes_IsComputedField(t *testing.T) {
	tests := []struct {
		name string
		args *args
		path []string
		want bool
	}{
		{
			"Is computed",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {
							Configshema: configschema.Attribute{
								Computed: true,
							},
						},
					},
				},
			},
			[]string{"test", "has", "tags"},
			true,
		},
		{
			"Not computed",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {Configshema: configschema.Attribute{
							Computed: false,
						}},
					},
				},
			},
			[]string{"test", "has", "tags"},
			false,
		},
		{
			"No tags",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {Configshema: configschema.Attribute{
							Computed: true,
						}},
					},
				},
			},
			[]string{"test", "has", "no", "tags"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := gocty.ImpliedType(tt.args.Attrs)
			ctyVal, _ := gocty.ToCtyValue(tt.args.Attrs, typ)

			attrs := CtyAttributes{
				Attrs:    tt.args.Attrs,
				value:    &ctyVal,
				metadata: tt.args.metadata,
			}
			if got := attrs.IsComputedField(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsComputedField() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCtyAttributes_IsJsonStringField(t *testing.T) {
	tests := []struct {
		name string
		args *args
		path []string
		want bool
	}{
		{
			"Is json",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {JsonString: true},
					},
				},
			},
			[]string{"test", "has", "tags"},
			true,
		},
		{
			"Not json",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {Configshema: configschema.Attribute{
							Computed: true,
						}},
					},
				},
			},
			[]string{"test", "has", "tags"},
			false,
		},
		{
			"No tags",
			&args{
				map[string]interface{}{
					"test": map[string]interface{}{
						"has": map[string]interface{}{
							"no": map[string]interface{}{
								"tags": "string",
							},
							"tags": "string",
						},
					},
				},
				&ResourceMetadata{
					AttributeMetadata: map[string]AttributeMetadata{
						"test.has.tags": {Configshema: configschema.Attribute{
							Computed: true,
						}},
					},
				},
			},
			[]string{"test", "has", "no", "tags"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, _ := gocty.ImpliedType(tt.args.Attrs)
			ctyVal, _ := gocty.ToCtyValue(tt.args.Attrs, typ)

			attrs := CtyAttributes{
				Attrs:    tt.args.Attrs,
				value:    &ctyVal,
				metadata: tt.args.metadata,
			}
			if got := attrs.IsJsonStringField(tt.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsComputedField() = %v, want %v", got, tt.want)
			}
		})
	}
}
