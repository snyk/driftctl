package state

import "fmt"

type StateReadingAlert struct {
	key string
	err string
}

func NewStateReadingAlert(key string, err error) *StateReadingAlert {
	return &StateReadingAlert{key: key, err: err.Error()}
}

func (s *StateReadingAlert) Message() string {
	return fmt.Sprintf("Your analysis will be incomplete. There was an error reading state file '%s': %s", s.key, s.err)
}

func (s *StateReadingAlert) ShouldIgnoreResource() bool {
	return false
}
