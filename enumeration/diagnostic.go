package enumeration

import "github.com/snyk/driftctl/enumeration/resource"

type Diagnostic interface {
	Code() string
	Message() string
	ResourceType() string
	Resource() *resource.Resource
}

type Diagnostics []Diagnostic
