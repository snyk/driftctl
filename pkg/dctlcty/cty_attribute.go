package dctlcty

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	if a.metadata == nil || a.Attrs == nil {
		return ""
	}

	var current interface{} = a.Attrs
	realPath := ""
	for _, part := range path {
		if current == nil {
			logrus.Debugf("Failed to find tag for path %+v", path)
			return ""
		}
		kind := reflect.TypeOf(current).Kind()
		switch kind {
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			index, err := strconv.ParseUint(part, 10, 64)
			if err != nil {
				logrus.Debugf("Failed to find tag for path %+v", path)
				return ""
			}
			current = current.([]interface{})[index]
			continue
		case reflect.Map:
			current = current.(map[string]interface{})[part]
		default:
			logrus.Debugf("Failed to find tag for path %+v", path)
			return ""
		}
		if realPath != "" {
			realPath = fmt.Sprintf("%s.%s", realPath, part)
			continue
		}
		realPath = part
	}

	fieldTags, exists := a.metadata.tags[realPath]
	if !exists {
		return ""
	}

	return reflect.StructTag(fieldTags)
}

func (a *CtyAttributes) IsComputedField(path []string) bool {
	tags := a.Tags(path)
	return tags.Get("computed") == "true"
}

func (a *CtyAttributes) IsJsonStringField(path []string) bool {
	tags := a.Tags(path)
	return tags.Get("jsonstring") == "true"
}
