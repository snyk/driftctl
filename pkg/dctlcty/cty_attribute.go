package dctlcty

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

func AsAttrs(val *cty.Value) CtyAttributes {
	if val == nil {
		return nil
	}
	bytes, _ := ctyjson.Marshal(*val, val.Type())
	var attrs map[string]interface{}
	err := json.Unmarshal(bytes, &attrs)
	if err != nil {
		panic(err)
	}
	return attrs
}

type CtyAttributes map[string]interface{}

func (a *CtyAttributes) SafeDelete(path []string) {
	val := *a
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
	val := *a
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
