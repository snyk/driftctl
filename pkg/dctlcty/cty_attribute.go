package dctlcty

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func AsAttrs(val *cty.Value, terraformType string) *CtyAttributes {
	if val == nil {
		return nil
	}

	metadata := resourcesMetadata[terraformType]

	bytes, _ := ctyjson.Marshal(*val, val.Type())
	var attrs map[string]interface{}
	err := json.Unmarshal(bytes, &attrs)
	if err != nil {
		panic(err)
	}

	attributes := &CtyAttributes{
		attrs,
		val,
		metadata,
	}

	if metadata != nil && metadata.Normalizer != nil {
		metadata.Normalizer(attributes)
	}

	return attributes
}

type CtyAttributes struct {
	Attrs    map[string]interface{}
	value    *cty.Value
	metadata *ResourceMetadata
}

func (a *CtyAttributes) SafeDelete(path []string) {
	val := a.Attrs
	for i, key := range path {
		if i == len(path)-1 {
			delete(val, key)
			return
		}

		v, exists := val[key]
		if !exists {
			return
		}
		m, ok := v.(map[string]interface{})
		if !ok {
			return
		}
		val = m
	}
}

func (a *CtyAttributes) SafeSet(path []string, value interface{}) error {
	val := a.Attrs
	for i, key := range path {
		if i == len(path)-1 {
			val[key] = value
			return nil
		}

		v, exists := val[key]
		if !exists {
			val[key] = map[string]interface{}{}
			v = val[key]
		}

		m, ok := v.(map[string]interface{})
		if !ok {
			return errors.Errorf("Path %s cannot be set: %s is not a nested struct", strings.Join(path, "."), key)
		}
		val = m
	}
	return errors.New("Error setting value") // should not happen ?
}

func (a *CtyAttributes) AttributeMetadata(path []string) *AttributeMetadata {
	if a.metadata == nil || a.value == nil {
		return nil
	}

	currentType := a.value.Type()
	var realPath []string
	for i, part := range path {

		if currentType.IsListType() || currentType.IsSetType() || currentType.IsTupleType() {
			currentType = currentType.ElementType()
			continue
		}

		if currentType.IsCollectionType() {
			currentType = currentType.ElementType()
		}

		if currentType.IsObjectType() {
			if !currentType.HasAttribute(part) {
				return nil // path doest not match this object
			}
			currentType = currentType.AttributeType(part)
		}

		if currentType.IsPrimitiveType() {
			if i < len(path)-1 {
				return nil // path leads to a non existing field
			}
		}
		realPath = append(realPath, part)
	}

	attributeMetadata, exists := a.metadata.AttributeMetadata[strings.Join(realPath, ".")]
	if !exists {
		return nil
	}

	return &attributeMetadata
}

func (a *CtyAttributes) IsComputedField(path []string) bool {
	metadata := a.AttributeMetadata(path)
	if metadata == nil {
		return false
	}
	return metadata.Configshema.Computed
}

func (a *CtyAttributes) IsJsonStringField(path []string) bool {
	metadata := a.AttributeMetadata(path)
	if metadata == nil {
		return false
	}
	return metadata.JsonString
}
