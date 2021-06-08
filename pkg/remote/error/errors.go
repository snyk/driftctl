package error

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/resource"
)

type SupplierError struct {
	err          error
	context      map[string]string
	supplierType string
}

func NewSupplierError(err error, context map[string]string, supplierType string) *SupplierError {
	context["SupplierType"] = supplierType
	return &SupplierError{err: err, context: context, supplierType: supplierType}
}

func (b *SupplierError) Error() string {
	return fmt.Sprintf("error in supplier %s: %s", b.supplierType, b.err)
}

func (b *SupplierError) RootCause() error {
	return b.err
}

func (b *SupplierError) SupplierType() string {
	return b.supplierType
}

func (b *SupplierError) Context() map[string]string {
	return b.context
}

type ResourceEnumerationError struct {
	SupplierError
	listedTypeError string
}

func NewResourceEnumerationErrorWithType(error error, supplierType resource.ResourceType, listedTypeError resource.ResourceType) *ResourceEnumerationError {
	context := map[string]string{
		"ListedTypeError": listedTypeError.String(),
	}
	return &ResourceEnumerationError{
		SupplierError:   *NewSupplierError(error, context, supplierType.String()),
		listedTypeError: listedTypeError.String(),
	}
}

func NewResourceEnumerationError(error error, supplierType resource.ResourceType) *ResourceEnumerationError {
	return NewResourceEnumerationErrorWithType(error, supplierType, supplierType)
}

func (b *ResourceEnumerationError) ListedTypeError() string {
	return b.listedTypeError
}
