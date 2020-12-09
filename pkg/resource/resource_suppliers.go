package resource

var resourceSupplier = make([]Supplier, 0)

func AddSupplier(supplier Supplier) {
	resourceSupplier = append(resourceSupplier, supplier)
}

func Suppliers() []Supplier {
	return resourceSupplier
}
