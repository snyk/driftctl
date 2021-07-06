package remote

import (
	"context"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/parallel"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ScannerOptions struct {
	Deep bool
}

type Scanner struct {
	resourceSuppliers    []resource.Supplier
	enumeratorRunner     *parallel.ParallelRunner
	detailsFetcherRunner *parallel.ParallelRunner
	remoteLibrary        *common.RemoteLibrary
	alerter              alerter.AlerterInterface
	options              ScannerOptions
}

func NewScanner(resourceSuppliers []resource.Supplier, remoteLibrary *common.RemoteLibrary, alerter alerter.AlerterInterface, options ScannerOptions) *Scanner {
	return &Scanner{
		resourceSuppliers:    resourceSuppliers,
		enumeratorRunner:     parallel.NewParallelRunner(context.TODO(), 10),
		detailsFetcherRunner: parallel.NewParallelRunner(context.TODO(), 10),
		remoteLibrary:        remoteLibrary,
		alerter:              alerter,
		options:              options,
	}
}

func (s *Scanner) retrieveRunnerResults(runner *parallel.ParallelRunner) ([]resource.Resource, error) {
	results := make([]resource.Resource, 0)
loop:
	for {
		select {
		case resources, ok := <-runner.Read():
			if !ok || resources == nil {
				break loop
			}

			for _, res := range resources.([]resource.Resource) {
				if res != nil {
					results = append(results, res)
				}
			}
		case <-runner.DoneChan():
			break loop
		}
	}
	return results, runner.Err()
}

func (s *Scanner) legacyScan() ([]resource.Resource, error) {
	for _, resourceProvider := range s.resourceSuppliers {
		supplier := resourceProvider
		s.enumeratorRunner.Run(func() (interface{}, error) {
			res, err := supplier.Resources()
			if err != nil {
				err := HandleResourceEnumerationError(err, s.alerter)
				if err == nil {
					return []resource.Resource{}, nil
				}
				return nil, err
			}
			for _, resource := range res {
				logrus.WithFields(logrus.Fields{
					"id":   resource.TerraformId(),
					"type": resource.TerraformType(),
				}).Debug("[DEPRECATED] Found cloud resource")
			}
			return res, nil
		})
	}

	return s.retrieveRunnerResults(s.enumeratorRunner)
}

func (s *Scanner) scan() ([]resource.Resource, error) {
	for _, enumerator := range s.remoteLibrary.Enumerators() {
		enumerator := enumerator
		s.enumeratorRunner.Run(func() (interface{}, error) {
			resources, err := enumerator.Enumerate()
			if err != nil {
				err := HandleResourceEnumerationError(err, s.alerter)
				if err == nil {
					return []resource.Resource{}, nil
				}
				return nil, err
			}
			for _, resource := range resources {
				if resource == nil {
					continue
				}
				logrus.WithFields(logrus.Fields{
					"id":   resource.TerraformId(),
					"type": resource.TerraformType(),
				}).Debug("Found cloud resource")
			}
			return resources, nil
		})
	}

	enumerationResult, err := s.retrieveRunnerResults(s.enumeratorRunner)
	if err != nil {
		return nil, err
	}

	for _, res := range enumerationResult {
		res := res
		s.detailsFetcherRunner.Run(func() (interface{}, error) {
			fetcher := s.remoteLibrary.GetDetailsFetcher(resource.ResourceType(res.TerraformType()))
			if fetcher != nil {
				// If we are in deep mode, retrieve resource details
				if s.options.Deep {
					resourceWithDetails, err := fetcher.ReadDetails(res)
					if err != nil {
						return nil, err
					}
					return []resource.Resource{resourceWithDetails}, nil
				}
			}
			return []resource.Resource{res}, nil
		})
	}

	return s.retrieveRunnerResults(s.detailsFetcherRunner)
}

func (s *Scanner) Resources() ([]resource.Resource, error) {

	resources, err := s.legacyScan()
	if err != nil {
		return nil, err
	}

	s.enumeratorRunner = parallel.NewParallelRunner(context.TODO(), 10)

	enumerationResult, err := s.scan()
	if err != nil {
		return nil, err
	}
	resources = append(resources, enumerationResult...)

	return resources, err
}

func (s *Scanner) Stop() {
	logrus.Debug("Stopping scanner")
	s.enumeratorRunner.Stop(errors.New("interrupted"))
	s.detailsFetcherRunner.Stop(errors.New("interrupted"))
}
