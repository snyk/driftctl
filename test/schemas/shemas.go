package schemas

import (
	"embed"
	gojson "encoding/json"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/terraform/providers"
)

//go:embed */*/schema.json
var fakeSchemaFS embed.FS

func WriteTestSchema(schema map[string]providers.Schema, provider, version string) error {
	_, relativeFilePath, _, _ := runtime.Caller(0)
	fileName := path.Join(path.Dir(relativeFilePath), provider, version, "schema.json")
	content, _ := gojson.Marshal(schema)
	err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, content, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

func ReadTestSchema(provider, version string) (map[string]providers.Schema, error) {
	content, err := fakeSchemaFS.ReadFile(path.Join(provider, version, "schema.json"))
	if err != nil {
		return nil, err
	}
	var schema map[string]providers.Schema
	if err := gojson.Unmarshal(content, &schema); err != nil {
		return nil, err
	}
	return schema, nil
}
