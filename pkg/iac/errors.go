package iac

import (
	"fmt"
	"strings"
)

type StateReadingError struct {
	errors []error
}

func NewStateReadingError() *StateReadingError {
	return &StateReadingError{}
}

func (s *StateReadingError) Add(err error) {
	s.errors = append(s.errors, err)
}

func (s *StateReadingError) Error() string {
	var err strings.Builder
	_, _ = fmt.Fprint(&err, "There were errors reading your states files : \n")
	for _, e := range s.errors {
		_, _ = fmt.Fprintf(&err, "   - %s\n", e.Error())
	}
	return err.String()
}
