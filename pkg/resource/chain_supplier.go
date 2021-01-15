package resource

import (
	"context"
	"runtime"

	"github.com/cloudskiff/driftctl/pkg/parallel"
)

type ChainSupplier struct {
	suppliers []Supplier
	runner    *parallel.ParallelRunner
}

func NewChainSupplier() *ChainSupplier {
	return &ChainSupplier{
		runner: parallel.NewParallelRunner(context.TODO(), int64(runtime.NumCPU())),
	}
}

func (r *ChainSupplier) AddSupplier(supplier Supplier) {
	r.suppliers = append(r.suppliers, supplier)
}

func (r *ChainSupplier) Resources() ([]Resource, error) {

	for _, supplier := range r.suppliers {
		sup := supplier
		r.runner.Run(func() (interface{}, error) {
			return sup.Resources()
		})
	}

	results := make([]Resource, 0)

ReadLoop:
	for {
		select {
		case supplierResult, ok := <-r.runner.Read():
			if !ok || supplierResult == nil {
				break ReadLoop
			}
			// Type cannot be invalid as return type is enforced
			// by Supplier interface
			resources, _ := supplierResult.([]Resource)
			results = append(results, resources...)
		case <-r.runner.DoneChan():
			break ReadLoop
		}
	}

	if r.runner.Err() != nil {
		return nil, r.runner.Err()
	}

	return results, nil
}
