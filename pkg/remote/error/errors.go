package error

import "fmt"

type RemoteError interface {
	ListedTypeError() string
}

type ResourceScanningError struct {
	err             error
	resourceType    string
	resourceId      string
	listedTypeError string
}

func (b *ResourceScanningError) Error() string {
	if b.resourceId != "" {
		return fmt.Sprintf("error scanning resource %s: %s", b.Resource(), b.err)
	}
	return fmt.Sprintf("error scanning resource type %s: %s", b.Resource(), b.err)
}

func (b *ResourceScanningError) RootCause() error {
	return b.err
}

func (b *ResourceScanningError) ResourceType() string {
	return b.resourceType
}

func NewResourceScanningError(error error, resourceType string, resourceId string) *ResourceScanningError {
	return &ResourceScanningError{
		err:             error,
		resourceType:    resourceType,
		resourceId:      resourceId,
		listedTypeError: resourceType,
	}
}

func NewResourceListingError(error error, resourceType string) *ResourceScanningError {
	return NewResourceListingErrorWithType(error, resourceType, resourceType)
}

func NewResourceListingErrorWithType(error error, resourceType, listedTypeError string) *ResourceScanningError {
	return &ResourceScanningError{
		err:             error,
		resourceType:    resourceType,
		listedTypeError: listedTypeError,
	}
}

func (b *ResourceScanningError) ListedTypeError() string {
	return b.listedTypeError
}

func (b *ResourceScanningError) Resource() string {
	if b.resourceId != "" {
		return fmt.Sprintf("%s.%s", b.resourceType, b.resourceId)
	}
	return b.resourceType
}

func (b *ResourceScanningError) String() string {
	return fmt.Sprintf("%s.%s (%s)", b.resourceType, b.resourceId, b.listedTypeError)
}
