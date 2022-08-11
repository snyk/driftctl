package resource

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type Source interface {
	Source() string
	Namespace() string
	InternalName() string
}

type SerializableSource struct {
	S    string `json:"source"`
	Ns   string `json:"namespace"`
	Name string `json:"internal_name"`
}

type TerraformStateSource struct {
	State  string
	Module string
	Name   string
}

func NewTerraformStateSource(state, module, name string) *TerraformStateSource {
	return &TerraformStateSource{state, module, name}
}

func (s *TerraformStateSource) Source() string {
	return s.State
}

func (s *TerraformStateSource) Namespace() string {
	return s.Module
}

func (s *TerraformStateSource) InternalName() string {
	return s.Name
}

type Resource struct {
	Id     string
	Type   string
	Attrs  *Attributes
	Sch    *Schema `json:"-" diff:"-"`
	Source Source  `json:"-"`
}

func (r *Resource) Schema() *Schema {
	return r.Sch
}

func (r *Resource) ResourceId() string {
	return r.Id
}

func (r *Resource) ResourceType() string {
	return r.Type
}

func (r *Resource) Attributes() *Attributes {
	return r.Attrs
}

func (r *Resource) Src() Source {
	return r.Source
}

func (r *Resource) SourceString() string {
	if r.Source == nil {
		return ""
	}
	if r.Source.Namespace() == "" {
		return fmt.Sprintf("%s.%s", r.ResourceType(), r.Source.InternalName())
	}
	return fmt.Sprintf("%s.%s.%s", r.Source.Namespace(), r.ResourceType(), r.Source.InternalName())
}

func (r *Resource) Equal(res *Resource) bool {
	if r.ResourceId() != res.ResourceId() || r.ResourceType() != res.ResourceType() {
		return false
	}

	if r.Schema() != nil && r.Schema().DiscriminantFunc != nil {
		return r.Schema().DiscriminantFunc(r, res)
	}

	return true
}

type ResourceFactory interface {
	CreateAbstractResource(ty, id string, data map[string]interface{}) *Resource
}

type SerializableResource struct {
	Id                 string              `json:"id"`
	Type               string              `json:"type"`
	ReadableAttributes map[string]string   `json:"human_readable_attributes,omitempty"`
	Source             *SerializableSource `json:"source,omitempty"`
}

func NewSerializableResource(res *Resource) *SerializableResource {
	var src *SerializableSource
	if res.Src() != nil {
		src = &SerializableSource{
			S:    res.Src().Source(),
			Ns:   res.Src().Namespace(),
			Name: res.Src().InternalName(),
		}
	}
	return &SerializableResource{
		Id:                 res.ResourceId(),
		Type:               res.ResourceType(),
		ReadableAttributes: formatReadableAttributes(res),
		Source:             src,
	}
}

func formatReadableAttributes(res *Resource) map[string]string {
	if res.Schema() == nil || res.Schema().HumanReadableAttributesFunc == nil {
		return map[string]string{}
	}
	return res.Schema().HumanReadableAttributesFunc(res)
}

type NormalizedResource interface {
	NormalizeForState() (Resource, error)
	NormalizeForProvider() (Resource, error)
}

func Sort(res []*Resource) []*Resource {
	sort.SliceStable(res, func(i, j int) bool {
		if res[i].ResourceType() != res[j].ResourceType() {
			return res[i].ResourceType() < res[j].ResourceType()
		}
		return res[i].ResourceId() < res[j].ResourceId()
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
	val, exist := (*a)[path]
	if !exist {
		return nil
	}
	if v, isInt := val.(int); isInt {
		return &v
	}
	floatVal := a.GetFloat64(path)
	if val == nil {
		return nil
	}
	v := int(*floatVal)
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
	attributesCopy := reflect.New(original.Type()).Elem()
	a.sanitize("", original, attributesCopy)
	*a = attributesCopy.Interface().(Attributes)
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
