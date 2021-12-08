package remote

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/parallel"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/resource"
)

type ScannerOptions struct {
	Deep bool
}

type Scanner struct {
	enumeratorRunner     *parallel.ParallelRunner
	detailsFetcherRunner *parallel.ParallelRunner
	remoteLibrary        *common.RemoteLibrary
	alerter              alerter.AlerterInterface
	options              ScannerOptions
	filter               filter.Filter
}

func NewScanner(remoteLibrary *common.RemoteLibrary, alerter alerter.AlerterInterface, options ScannerOptions, filter filter.Filter) *Scanner {
	return &Scanner{
		enumeratorRunner:     parallel.NewParallelRunner(context.TODO(), 10),
		detailsFetcherRunner: parallel.NewParallelRunner(context.TODO(), 10),
		remoteLibrary:        remoteLibrary,
		alerter:              alerter,
		options:              options,
		filter:               filter,
	}
}

func (s *Scanner) retrieveRunnerResults(runner *parallel.ParallelRunner) ([]*resource.Resource, error) {
	results := make([]*resource.Resource, 0)
loop:
	for {
		select {
		case resources, ok := <-runner.Read():
			if !ok || resources == nil {
				break loop
			}

			for _, res := range resources.([]*resource.Resource) {
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

func (s *Scanner) EnumerateResources() ([]*resource.Resource, error) {
	for _, enumerator := range s.remoteLibrary.Enumerators() {
		if s.filter.IsTypeIgnored(enumerator.SupportedType()) {
			logrus.WithFields(logrus.Fields{
				"type": enumerator.SupportedType(),
			}).Debug("Ignored enumeration of resources since it is ignored in filter")
			continue
		}
		enumerator := enumerator
		s.enumeratorRunner.Run(func() (interface{}, error) {
			resources, err := enumerator.Enumerate()
			if err != nil {
				err := HandleResourceEnumerationError(err, s.alerter)
				if err == nil {
					return []*resource.Resource{}, nil
				}
				return nil, err
			}
			for _, res := range resources {
				if res == nil {
					continue
				}
				logrus.WithFields(logrus.Fields{
					"id":   res.ResourceId(),
					"type": res.ResourceType(),
				}).Debug("Found cloud resource")
			}
			return resources, nil
		})
	}

	return s.retrieveRunnerResults(s.enumeratorRunner)
}

func (s *Scanner) ReadResources(managedResources []*resource.Resource) ([]*resource.Resource, error) {
	for _, res := range managedResources {
		res := res
		s.detailsFetcherRunner.Run(func() (interface{}, error) {
			fetcher := s.remoteLibrary.GetDetailsFetcher(resource.ResourceType(res.ResourceType()))
			if fetcher == nil {
				return []*resource.Resource{res}, nil
			}

			resourceWithDetails, err := fetcher.ReadDetails(res)
			if err != nil {
				if err := HandleResourceDetailsFetchingError(err, s.alerter); err != nil {
					return nil, err
				}
				return []*resource.Resource{}, nil
			}
			return []*resource.Resource{resourceWithDetails}, nil
		})
	}

	return s.retrieveRunnerResults(s.detailsFetcherRunner)
}

func (s *Scanner) Resources() ([]*resource.Resource, error) {
	resources, err := s.EnumerateResources()
	if err != nil {
		return nil, err
	}

	if !s.options.Deep {
		return resources, nil
	}

	// Be aware that this call will read all resources, no matter they're managed or not
	resources, err = s.ReadResources(resources)
	if err != nil {
		return nil, err
	}
	return resources, err
}

func (s *Scanner) Stop() {
	logrus.Debug("Stopping scanner")
	s.enumeratorRunner.Stop(errors.New("interrupted"))
	s.detailsFetcherRunner.Stop(errors.New("interrupted"))
}
