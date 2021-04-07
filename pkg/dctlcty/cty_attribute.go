package dctlcty

import (
	"encoding/json"
	"reflect"
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
		&metadata,
	}

	if metadata.normalizer != nil {
		metadata.normalizer(attributes)
	}

	return attributes
}

type CtyAttributes struct {
	Attrs    map[string]interface{}
	metadata *Metadata
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

func (a *CtyAttributes) Tags(path []string) reflect.StructTag {
	if a.metadata == nil {
		return ""
	}

	fieldTags, exists := a.metadata.tags[strings.Join(path, ".")]
	if !exists {
		return ""
	}

	return reflect.StructTag(fieldTags)
}

func (a *CtyAttributes) IsComputedField(path []string) bool {
	tags := a.Tags(path)
	return tags.Get("computed") == "true"
}
