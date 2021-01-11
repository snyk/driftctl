package resource

import "fmt"

type FakeResource struct {
	Id        string
	FooBar    string
	BarFoo    string `computed:"true"`
	Json      string `jsonstring:"true"`
	Type      string
	Tags      map[string]string
	CustomMap map[string]struct {
		Tag string
	}
	Slice  []string
	Struct struct {
		Baz string `computed:"true"`
		Bar string
	}
	StructSlice []struct {
		String string   `computed:"true"`
		Array  []string `computed:"true"`
	}
}

func (d FakeResource) TerraformId() string {
	return d.Id
}

func (d FakeResource) TerraformType() string {
	if d.Type != "" {
		return d.Type
	}
	return "FakeResource"
}

type FakeResourceStringer struct {
	Id   string
	Name string
}

func (d *FakeResourceStringer) TerraformId() string {
	return d.Id
}

func (d *FakeResourceStringer) TerraformType() string {
	return "FakeResourceStringer"
}

func (d *FakeResourceStringer) String() string {
	return fmt.Sprintf("Name: '%s'", d.Name)
}
