package supplier

import (
	"context"
	"runtime"

	"github.com/snyk/driftctl/pkg/iac"
	"github.com/snyk/driftctl/pkg/parallel"
	"github.com/snyk/driftctl/pkg/resource"
)

type IacChainSupplier struct {
	suppliers []resource.Supplier
	runner    *parallel.ParallelRunner
}

func NewIacChainSupplier() *IacChainSupplier {
	return &IacChainSupplier{
		runner: parallel.NewParallelRunner(context.TODO(), int64(runtime.NumCPU())),
	}
}

func (r *IacChainSupplier) AddSupplier(supplier resource.Supplier) {
	r.suppliers = append(r.suppliers, supplier)
}

func (r *IacChainSupplier) Resources() ([]*resource.Resource, error) {

	for _, supplier := range r.suppliers {
		sup := supplier
		r.runner.Run(func() (interface{}, error) {
			resources, err := sup.Resources()
			return &result{err, resources}, nil
		})
	}

	results := make([]*resource.Resource, 0)
	isSuccess := false
	retrieveError := iac.NewStateReadingError()
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
				retrieveError.Add(result.err)
				continue
			}
			isSuccess = true
			results = append(results, result.res...)
		case <-r.runner.DoneChan():
			break ReadLoop
		}
	}

	if r.runner.Err() != nil {
		return nil, r.runner.Err()
	}

	if !isSuccess {
		// only fail if all suppliers failed
		return nil, retrieveError
	}

	return results, nil
}

type result struct {
	err error
	res []*resource.Resource
}
