package deserializer

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/zclconf/go-cty/cty/gocty"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/zclconf/go-cty/cty"
)

type deserializer struct {
	normalize func(res resource.Resource) error
}

func (d *deserializer) deserialize(rawList []cty.Value, to interface{}) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, rawResource := range rawList {
		rawResource := rawResource

		if err := gocty.FromCtyValue(rawResource, to); err != nil {
			return nil, errors.Wrapf(err, "Cannot deserialize object of type %s", reflect.TypeOf(to).Name())
		}

		decoded := clone(to).(resource.Resource)
		if d.normalize != nil {
			if err := d.normalize(decoded); err != nil {
				return nil, err
			}
		}

		resources = append(resources, decoded)
	}
	return resources, nil
}

func clone(inter interface{}) interface{} {
	nInter := reflect.New(reflect.TypeOf(inter).Elem())

	val := reflect.ValueOf(inter).Elem()
	nVal := nInter.Elem()
	for i := 0; i < val.NumField(); i++ {
		nvField := nVal.Field(i)
		nvField.Set(val.Field(i))
	}

	return nInter.Interface()
}
