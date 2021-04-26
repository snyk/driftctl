package schemas

import (
	gojson "encoding/json"
	"io/ioutil"
	"path"
	"runtime"

	"github.com/hashicorp/terraform/providers"
)

func ReadTestSchema(provider, version string) (map[string]providers.Schema, error) {
	_, filename, _, _ := runtime.Caller(0)
	content, err := ioutil.ReadFile(path.Join(path.Dir(filename), provider, version, "schema.json"))
	if err != nil {
		return nil, err
	}
	var schema map[string]providers.Schema
	if err := gojson.Unmarshal(content, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}
