package resource

type SupplierLibrary struct {
	resourceSupplier []Supplier
}

func NewSupplierLibrary() *SupplierLibrary {
	return &SupplierLibrary{
		make([]Supplier, 0),
	}
}

func (r *SupplierLibrary) AddSupplier(supplier Supplier) {
	r.resourceSupplier = append(r.resourceSupplier, supplier)
}

func (r *SupplierLibrary) Suppliers() []Supplier {
	return r.resourceSupplier
}
