package resource

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Resource interface {
	TerraformId() string
	TerraformType() string
	TerraformImportId() string
	Attributes() *Attributes
	Schema() *Schema
}

type AbstractResource struct {
	Id    string
	Type  string
	Attrs *Attributes
	Sch   *Schema `json:"-" diff:"-"`
}

func (a *AbstractResource) Schema() *Schema {
	return a.Sch
}

func (a *AbstractResource) TerraformId() string {
	return a.Id
}

func (a *AbstractResource) TerraformType() string {
	return a.Type
}

func (a *AbstractResource) TerraformImportId() string {
	// NOTE: Sometime just id, sometime more sophisticated string / ARN, check by type
	importId := ""
	switch a.Type {
	case "aws_vpc":
		importId = a.Id
	case "aws_subnet":
		importId = a.Id
	case "aws_sns_topic_subscription":
		importId = a.Id
	case "aws_s3_bucket_policy":
		importId = a.Id
	case "aws_iam_access_key":
		importId = a.Id
	case "aws_key_pair":
		importId = a.Id
	case "aws_iam_policy":
		importId = a.Id
	case "aws_iam_role":
		importId = a.Id
	case "aws_iam_role_policy":
		importId = a.Id
	case "aws_iam_user":
		importId = a.Id
	case "aws_s3_bucket":
		importId = a.Id
	case "aws_security_group":
		importId = a.Id
	case "aws_iam_policy_attachment":
		// NOTE: terraform not supporting as well
		// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/iam_policy_attachment
		importId = ""
	case "aws_security_group_rule":
		importId = ""
		// TODO: similar to CreateSecurityGroupRuleIdHash , need some custom type handlings maybe
		// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/security_group_rule#import
	case "aws_route":
		// NOTE:
		// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/route
		// One of the following destination arguments must be supplied:
		// destination_cidr_block - (Optional) The destination CIDR block.
		// destination_ipv6_cidr_block - (Optional) The destination IPv6 CIDR block.
		// destination_prefix_list_id - (Optional) The ID of a managed prefix list destination.
		_, okDest1 := a.Attrs.Get("destination_cidr_block")
		_, okDest2 := a.Attrs.Get("destination_ipv6_cidr_block")
		dest := ""
		if okDest1 {
			dest = *a.Attrs.GetString("destination_cidr_block")
		} else if okDest2 {
			dest = *a.Attrs.GetString("destination_ipv6_cidr_block")
		} else {
			dest = *a.Attrs.GetString("destination_prefix_list_id")
		}
		importId = fmt.Sprintf("%s_%s", *a.Attrs.GetString("route_table_id"), dest)
	default:
		importId = ""
	}

	return importId
}

func (a *AbstractResource) Attributes() *Attributes {
	return a.Attrs
}

type ResourceFactory interface {
	CreateAbstractResource(ty, id string, data map[string]interface{}) *AbstractResource
}

type SerializableResource struct {
	Resource
}

type SerializedResource struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

func (u *SerializedResource) TerraformId() string {
	return u.Id
}

func (u *SerializedResource) TerraformType() string {
	return u.Type
}

func (u *SerializedResource) TerraformImportId() string {
	return ""
}

func (u *SerializedResource) Attributes() *Attributes {
	return nil
}

func (u *SerializedResource) Schema() *Schema {
	return nil
}

func (s *SerializableResource) UnmarshalJSON(bytes []byte) error {
	var res *SerializedResource

	if err := json.Unmarshal(bytes, &res); err != nil {
		return err
	}
	s.Resource = res
	return nil
}

func (s SerializableResource) MarshalJSON() ([]byte, error) {
	return json.Marshal(SerializedResource{Id: s.TerraformId(), Type: s.TerraformType()})
}

type NormalizedResource interface {
	NormalizeForState() (Resource, error)
	NormalizeForProvider() (Resource, error)
}

func IsSameResource(rRs, lRs Resource) bool {
	return rRs.TerraformType() == lRs.TerraformType() && rRs.TerraformId() == lRs.TerraformId()
}

func Sort(res []Resource) []Resource {
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].TerraformType() != res[j].TerraformType() {
			return res[i].TerraformType() < res[j].TerraformType()
		}
		return res[i].TerraformId() < res[j].TerraformId()
	})
	return res
}

type Attributes map[string]interface{}

func (a *Attributes) Copy() *Attributes {
	res := Attributes{}

	for key, value := range *a {
		_ = res.SafeSet([]string{key}, value)
	}

	return &res
}

func (a *Attributes) Get(path string) (interface{}, bool) {
	val, exist := (*a)[path]
	return val, exist
}

func (a *Attributes) GetSlice(path string) []interface{} {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	return val.([]interface{})
}

func (a *Attributes) GetString(path string) *string {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	v := val.(string)
	return &v
}

func (a *Attributes) GetBool(path string) *bool {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	v := val.(bool)
	return &v
}

func (a *Attributes) GetInt(path string) *int {
	val := a.GetFloat64(path)
	if val == nil {
		return nil
	}
	v := int(*val)
	return &v
}

func (a *Attributes) GetFloat64(path string) *float64 {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	v := val.(float64)
	return &v
}

func (a *Attributes) GetMap(path string) map[string]interface{} {
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	return val.(map[string]interface{})
}

func (a *Attributes) SafeDelete(path []string) {
	for i, key := range path {
		if i == len(path)-1 {
			delete(*a, key)
			return
		}

		v, exists := (*a)[key]
		if !exists {
			return
		}
		m, ok := v.(Attributes)
		if !ok {
			return
		}
		*a = m
	}
}

func (a *Attributes) SafeSet(path []string, value interface{}) error {
	for i, key := range path {
		if i == len(path)-1 {
			(*a)[key] = value
			return nil
		}

		v, exists := (*a)[key]
		if !exists {
			(*a)[key] = map[string]interface{}{}
			v = (*a)[key]
		}

		m, ok := v.(Attributes)
		if !ok {
			return errors.Errorf("Path %s cannot be set: %s is not a nested struct", strings.Join(path, "."), key)
		}
		*a = m
	}
	return errors.New("Error setting value") // should not happen ?
}

func (a *Attributes) DeleteIfDefault(path string) {
	val, exist := a.Get(path)
	ty := reflect.TypeOf(val)
	if exist && val == reflect.Zero(ty).Interface() {
		a.SafeDelete([]string{path})
	}
}

func concatenatePath(path, next string) string {
	if path == "" {
		return next
	}
	return strings.Join([]string{path, next}, ".")
}

func (a *Attributes) SanitizeDefaults() {
	original := reflect.ValueOf(*a)
	copy := reflect.New(original.Type()).Elem()
	a.sanitize("", original, copy)
	*a = copy.Interface().(Attributes)
}

func (a *Attributes) sanitize(path string, original, copy reflect.Value) bool {
	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return false
		}
		copy.Set(reflect.New(originalValue.Type()))
		a.sanitize(path, originalValue, copy.Elem())
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return false
		}
		if originalValue.Kind() == reflect.Slice || originalValue.Kind() == reflect.Map {
			if originalValue.Len() == 0 {
				return false
			}
		}
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		a.sanitize(path, originalValue, copyValue)
		copy.Set(copyValue)

	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			field := original.Field(i)
			a.sanitize(concatenatePath(path, field.String()), field, copy.Field(i))
		}
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			a.sanitize(concatenatePath(path, strconv.Itoa(i)), original.Index(i), copy.Index(i))
		}
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			created := a.sanitize(concatenatePath(path, key.String()), originalValue, copyValue)
			if created {
				copy.SetMapIndex(key, copyValue)
			}
		}
	default:
		copy.Set(original)
	}
	return true
}
