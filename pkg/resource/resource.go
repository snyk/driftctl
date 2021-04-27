package resource

import (
	"encoding/json"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Resource interface {
	TerraformId() string
	TerraformType() string
	CtyValue() *cty.Value
}

var refactoredResources = []string{
	"aws_ami",
	"aws_cloudfront_distribution",
	"aws_db_instance",
	"aws_db_subnet_group",
	"aws_default_route_table",
	"aws_default_security_group",
	"aws_default_subnet",
	"aws_default_vpc",
	"aws_dynamodb_table",
}

func IsRefactoredResource(typ string) bool {
	for _, refactoredResource := range refactoredResources {
		if typ == refactoredResource {
			return true
		}
	}
	return false
}

type AbstractResource struct {
	Id    string
	Type  string
	Attrs *Attributes
}

func (a *AbstractResource) TerraformId() string {
	return a.Id
}

func (a *AbstractResource) TerraformType() string {
	return a.Type
}

func (a *AbstractResource) CtyValue() *cty.Value {
	return nil
}

type ResourceFactory interface {
	CreateResource(data interface{}, ty string) (*cty.Value, error)
}

type SerializableResource struct {
	Resource
}

type SerializedResource struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

func (u SerializedResource) TerraformId() string {
	return u.Id
}

func (u SerializedResource) TerraformType() string {
	return u.Type
}

func (u SerializedResource) CtyValue() *cty.Value {
	return &cty.NilVal
}

func (s *SerializableResource) UnmarshalJSON(bytes []byte) error {
	var res SerializedResource

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

func ToResourceAttributes(val *cty.Value) *Attributes {
	if val == nil {
		return nil
	}

	bytes, _ := ctyjson.Marshal(*val, val.Type())
	var attrs Attributes
	err := json.Unmarshal(bytes, &attrs)
	if err != nil {
		panic(err)
	}

	return &attrs
}

type Attributes map[string]interface{}

func (a *Attributes) Get(path string) (interface{}, bool) {
	val, exist := (*a)[path]
	return val, exist
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

func (a *Attributes) SanitizeDefaults() {
	original := reflect.ValueOf(*a)
	copy := reflect.New(original.Type()).Elem()
	a.run("", original, copy)
	*a = copy.Interface().(Attributes)
}

func (a *Attributes) run(path string, original, copy reflect.Value) {
	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return
		}
		copy.Set(reflect.New(originalValue.Type()))
		a.run(path, originalValue, copy.Elem())
	case reflect.Interface:
		// Get rid of the wrapping interface
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return
		}
		if originalValue.Kind() == reflect.Slice || originalValue.Kind() == reflect.Map {
			if originalValue.Len() == 0 {
				return
			}
		}
		// Create a new object. Now new gives us a pointer, but we want the value it
		// points to, so we have to call Elem() to unwrap it
		copyValue := reflect.New(originalValue.Type()).Elem()
		a.run(path, originalValue, copyValue)
		copy.Set(copyValue)

	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			field := original.Field(i)
			a.run(concatenatePath(path, field.String()), field, copy.Field(i))
		}
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			a.run(concatenatePath(path, strconv.Itoa(i)), original.Index(i), copy.Index(i))
		}
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			a.run(concatenatePath(path, key.String()), originalValue, copyValue)
			copy.SetMapIndex(key, copyValue)
		}
	default:
		copy.Set(original)
	}
}

func concatenatePath(path, next string) string {
	if path == "" {
		return next
	}
	return strings.Join([]string{path, next}, ".")
}

func (a *Attributes) SanitizeDefaultsV3() {
	original := reflect.ValueOf(*a)
	copy := reflect.New(original.Type()).Elem()
	a.runV3("", original, copy)
	*a = copy.Interface().(Attributes)
}

func (a *Attributes) runV3(path string, original, copy reflect.Value) bool {
	switch original.Kind() {
	case reflect.Ptr:
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return false
		}
		copy.Set(reflect.New(originalValue.Type()))
		a.runV3(path, originalValue, copy.Elem())
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
		a.runV3(path, originalValue, copyValue)
		copy.Set(copyValue)

	case reflect.Struct:
		for i := 0; i < original.NumField(); i += 1 {
			field := original.Field(i)
			a.runV3(concatenatePath(path, field.String()), field, copy.Field(i))
		}
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i += 1 {
			a.runV3(concatenatePath(path, strconv.Itoa(i)), original.Index(i), copy.Index(i))
		}
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			created := a.runV3(concatenatePath(path, key.String()), originalValue, copyValue)
			if created {
				copy.SetMapIndex(key, copyValue)
			}
		}
	default:
		copy.Set(original)
	}
	return true
}
