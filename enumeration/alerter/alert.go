package alerter

import (
	"encoding/json"

	"github.com/snyk/driftctl/enumeration/resource"
)

type Alerts map[string][]Alert

type Alert interface {
	Message() string
	ShouldIgnoreResource() bool
	Resource() *resource.Resource
}

type FakeAlert struct {
	Msg            string
	IgnoreResource bool
}

func (f *FakeAlert) Message() string {
	return f.Msg
}

func (f *FakeAlert) ShouldIgnoreResource() bool {
	return f.IgnoreResource
}

func (f *FakeAlert) Resource() *resource.Resource {
	return nil
}

type SerializableAlert struct {
	Alert
}

type SerializedAlert struct {
	Msg string `json:"message"`
}

func (u *SerializedAlert) Message() string {
	return u.Msg
}

func (u *SerializedAlert) ShouldIgnoreResource() bool {
	return false
}

func (s *SerializedAlert) Resource() *resource.Resource {
	return nil
}

func (s *SerializableAlert) UnmarshalJSON(bytes []byte) error {
	var res SerializedAlert

	if err := json.Unmarshal(bytes, &res); err != nil {
		return err
	}
	s.Alert = &res
	return nil
}

func (s *SerializableAlert) MarshalJSON() ([]byte, error) {
	return json.Marshal(SerializedAlert{Msg: s.Message()})
}
