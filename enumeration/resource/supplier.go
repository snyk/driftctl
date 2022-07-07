package resource

// Supplier supply the list of resource.Resource, it's the main interface to retrieve remote resources
type Supplier interface {
	Resources() ([]*Resource, error)
}

type StoppableSupplier interface {
	Supplier
	Stop()
}
