package remote

import (
	"github.com/snyk/driftctl/enumeration/resource"
)

type SortableScanner struct {
	Scanner resource.Supplier
}

func NewSortableScanner(scanner resource.Supplier) *SortableScanner {
	return &SortableScanner{
		Scanner: scanner,
	}
}

func (s *SortableScanner) Resources() ([]*resource.Resource, error) {
	resources, err := s.Scanner.Resources()
	if err != nil {
		return nil, err
	}
	return resource.Sort(resources), nil
}
