package resource

// Supplier Resource Supplier supply the list of resource.Resource, its the front to retrieve remote resources
type Supplier interface {
	Resources() ([]Resource, error)
}

type SimpleTypeSupplier interface {
	Supplier
	SuppliedType() ResourceType
}

type StoppableSupplier interface {
	Supplier
	Stop()
}
