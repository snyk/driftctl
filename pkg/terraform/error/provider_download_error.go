package error

import "fmt"

type ProviderNotFoundError struct {
	Version string
}

func (p ProviderNotFoundError) Error() string {
	return fmt.Sprintf("Provider version %s does not exist", p.Version)
}
