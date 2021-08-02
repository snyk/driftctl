package error

import "fmt"

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

type ResourceScanningError struct {
	SupplierError
	listedTypeError string
}

func NewResourceScanningErrorWithType(error error, supplierType string, listedTypeError string) *ResourceScanningError {
	context := map[string]string{
		"ListedTypeError": listedTypeError,
	}
	return &ResourceScanningError{
		SupplierError:   *NewSupplierError(error, context, supplierType),
		listedTypeError: listedTypeError,
	}
}

func NewResourceScanningError(error error, supplierType string) *ResourceScanningError {
	return NewResourceScanningErrorWithType(error, supplierType, supplierType)
}

func (b *ResourceScanningError) ListedTypeError() string {
	return b.listedTypeError
}
