package helpers

import "encoding/json"

// Since we can't use both hashicorp/terraform and hashicorp/terraform-plugin-sdk
// dependencies together, we decided to duplicate the helper function below from
// the original repo.
// https://github.com/hashicorp/terraform-plugin-sdk/issues/268
// https://www.terraform.io/docs/extend/guides/v1-upgrade-guide.html
// https://github.com/hashicorp/terraform-website/blob/master/content/source/docs/extend/best-practices/depending-on-providers.html.md#using-the-rpc-protocol

// Takes a value containing JSON string and passes it through
// the JSON parser to normalize it, returns either a parsing
// error or normalized JSON string.
func NormalizeJsonString(jsonString interface{}) (string, error) {
	var j interface{}

	if jsonString == nil || jsonString.(string) == "" {
		return "", nil
	}

	s := jsonString.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return s, err
	}

	bytes, _ := json.Marshal(j)
	return string(bytes[:]), nil
}
