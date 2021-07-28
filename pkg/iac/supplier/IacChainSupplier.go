package supplier

import (
	"context"
	"runtime"

	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/pkg/errors"
)

type IacChainSupplier struct {
	suppliers []resource.Supplier
	runner    *parallel.ParallelRunner
}

func NewIacChainSupplier() *IacChainSupplier {
	return &IacChainSupplier{
		suppliers: []resource.Supplier{},
		runner:    parallel.NewParallelRunner(context.TODO(), int64(runtime.NumCPU())),
	}
}

func (r *IacChainSupplier) AddSupplier(supplier resource.Supplier) {
	r.suppliers = append(r.suppliers, supplier)
}

func (r *IacChainSupplier) CountSuppliers() int {
	return len(r.suppliers)
}

func (r *IacChainSupplier) Resources() ([]*resource.Resource, error) {

	if len(r.suppliers) <= 0 {
		return nil, errors.New("There was an error retrieving your states check alerts for details.")
	}

	for _, supplier := range r.suppliers {
		sup := supplier
		r.runner.Run(func() (interface{}, error) {
			resources, err := sup.Resources()
			return &result{err, resources}, nil
		})
	}

	results := make([]*resource.Resource, 0)
	nbErrors := 0
ReadLoop:
	for {
		select {
		case supplierResult, ok := <-r.runner.Read():
			if !ok || supplierResult == nil {
				break ReadLoop
			}
			// Type cannot be invalid as return type is enforced
			// in run function on top
			result, _ := supplierResult.(*result)

			if result.err != nil {
				nbErrors++
				continue
			}

			results = append(results, result.res...)
		case <-r.runner.DoneChan():
			break ReadLoop
		}
	}

	if r.runner.Err() != nil {
		return nil, r.runner.Err()
	}

	if nbErrors == len(r.suppliers) {
		// only fail if all suppliers failed
		return nil, errors.New("There was an error retrieving your states check alerts for details.")
	}

	return results, nil
}

type result struct {
	err error
	res []*resource.Resource
}
