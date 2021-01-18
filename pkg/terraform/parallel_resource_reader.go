package terraform

import (
	"github.com/cloudskiff/driftctl/pkg/parallel"

	"github.com/zclconf/go-cty/cty"
)

type ParallelResourceReader struct {
	runner *parallel.ParallelRunner
}

func NewParallelResourceReader(runner *parallel.ParallelRunner) *ParallelResourceReader {
	return &ParallelResourceReader{
		runner: runner,
	}
}

func (p *ParallelResourceReader) Wait() ([]cty.Value, error) {
	results := make([]cty.Value, 0)
Loop:
	for {
		select {
		case res, ok := <-p.runner.Read():
			if !ok {
				break Loop
			}
			ctyVal := res.(cty.Value)
			if !ctyVal.IsNull() {
				results = append(results, ctyVal)
			}
		case <-p.runner.DoneChan():
			break Loop
		}
	}
	return results, p.runner.Err()
}

func (p *ParallelResourceReader) Run(runnable func() (cty.Value, error)) {
	p.runner.Run(func() (interface{}, error) {
		return runnable()
	})
}
