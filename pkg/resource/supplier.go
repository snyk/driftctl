package resource

// Supplier supply the list of resource.Resource, it's the main interface to retrieve remote resources
type Supplier interface {
	Resources() ([]*Resource, error)
}

// IaCSupplier supply the list of resource.Resource, it's the main interface to retrieve state resources
type IaCSupplier interface {
	Supplier
	SourceCount() uint
}

type StoppableSupplier interface {
	Supplier
	Stop()
}
