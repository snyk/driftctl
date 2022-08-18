package enumerator

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/diagnostic"
	"github.com/snyk/driftctl/enumeration/parallel"
	"github.com/snyk/driftctl/enumeration/remote"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

type CloudEnumerator struct {
	alerter              *sliceAlerter
	progress             enumeration.ProgressCounter
	remoteLibrary        *common.RemoteLibrary
	providerLibrary      *terraform.ProviderLibrary
	enumeratorRunner     *parallel.ParallelRunner
	detailsFetcherRunner *parallel.ParallelRunner
	to                   string
}

type ListOutput struct {
	Resources   []*resource.Resource
	Diagnostics diagnostic.Diagnostics
}

type cloudEnumeratorBuilder struct {
	cloud           string
	providerVersion string
	configDirectory string
}

// WithCloud Choose which cloud to use for enumeration and refresh
// TODO could be inferred with types listed
func (b *cloudEnumeratorBuilder) WithCloud(cloud string) *cloudEnumeratorBuilder {
	b.cloud = cloud
	return b
}

// WithProviderVersion optionally choose the provider version used for refresh
func (b *cloudEnumeratorBuilder) WithProviderVersion(providerVersion string) *cloudEnumeratorBuilder {
	b.providerVersion = providerVersion
	return b
}

// WithConfigDirectory optionally choose the directory used to download terraform provider used for refresh
func (b *cloudEnumeratorBuilder) WithConfigDirectory(configDir string) *cloudEnumeratorBuilder {
	b.configDirectory = configDir
	return b
}

func (b *cloudEnumeratorBuilder) Build() (*CloudEnumerator, error) {
	enumerator := &CloudEnumerator{
		enumeratorRunner:     parallel.NewParallelRunner(context.TODO(), 10),
		detailsFetcherRunner: parallel.NewParallelRunner(context.TODO(), 10),
		providerLibrary:      terraform.NewProviderLibrary(),
		remoteLibrary:        common.NewRemoteLibrary(),
		alerter:              newSliceAlerter(),
		progress:             &dummyCounter{},
	}

	if b.configDirectory == "" {
		tempDir, err := os.MkdirTemp("", "enumerator")
		if err != nil {
			return nil, err
		}
		b.configDirectory = tempDir
	}

	err := enumerator.init(fmt.Sprintf("%s+tf", b.cloud), b.providerVersion, b.configDirectory)

	return enumerator, err
}

func NewCloudEnumerator() *cloudEnumeratorBuilder {
	return &cloudEnumeratorBuilder{}
}

func (e *CloudEnumerator) init(to, providerVersion, configDirectory string) error {
	e.to = to

	resFactory := terraform.NewTerraformResourceFactory()

	err := remote.Activate(to, providerVersion, e.alerter, e.providerLibrary, e.remoteLibrary, e.progress, resFactory, configDirectory)
	if err != nil {
		return err
	}
	return nil
}

func (e *CloudEnumerator) Enumerate(input *enumeration.EnumerateInput) (*enumeration.EnumerateOutput, error) {

	e.alerter.alerts = alerter.Alerts{}

	types := map[string]struct{}{}
	for _, resourceType := range input.ResourceTypes {
		types[resourceType] = struct{}{}
	}
	filter := typeFilter{types: types}

	for _, enumerator := range e.remoteLibrary.Enumerators() {
		if filter.IsTypeIgnored(enumerator.SupportedType()) {
			logrus.WithFields(logrus.Fields{
				"type": enumerator.SupportedType(),
			}).Debug("Ignored enumeration of resources since it is ignored in filter")
			continue
		}
		enumerator := enumerator
		e.enumeratorRunner.Run(func() (interface{}, error) {
			resources, err := enumerator.Enumerate()
			if err != nil {
				err := remote.HandleResourceEnumerationError(err, e.alerter)
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

	results, err := e.retrieveRunnerResults(e.enumeratorRunner)
	if err != nil {
		return nil, err
	}

	mapRes := mapByType(results)

	diagnostics := diagnostic.FromAlerts(e.alerter.Alerts())

	return &enumeration.EnumerateOutput{
		Resources:   mapRes,
		Timings:     nil,
		Diagnostics: diagnostics,
	}, nil
}

func (e *CloudEnumerator) Refresh(input *enumeration.RefreshInput) (*enumeration.RefreshOutput, error) {

	e.alerter.alerts = alerter.Alerts{}

	for _, resByType := range input.Resources {
		for _, res := range resByType {
			res := res
			e.detailsFetcherRunner.Run(func() (interface{}, error) {
				fetcher := e.remoteLibrary.GetDetailsFetcher(resource.ResourceType(res.ResourceType()))
				if fetcher == nil {
					return []*resource.Resource{res}, nil
				}

				resourceWithDetails, err := fetcher.ReadDetails(res)
				if err != nil {
					if err := remote.HandleResourceDetailsFetchingError(err, e.alerter); err != nil {
						return nil, err
					}
					return []*resource.Resource{}, nil
				}
				return []*resource.Resource{resourceWithDetails}, nil
			})
		}
	}

	results, err := e.retrieveRunnerResults(e.detailsFetcherRunner)
	if err != nil {
		return nil, err
	}

	mapRes := mapByType(results)
	diagnostics := diagnostic.FromAlerts(e.alerter.Alerts())

	return &enumeration.RefreshOutput{
		Resources:   mapRes,
		Diagnostics: diagnostics,
	}, nil
}

func (e *CloudEnumerator) GetSchema() (*enumeration.GetSchemasOutput, error) {
	panic("GetSchema is not implemented..")
}

func (e *CloudEnumerator) retrieveRunnerResults(runner *parallel.ParallelRunner) ([]*resource.Resource, error) {
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

func (e *CloudEnumerator) List(typ string) (*ListOutput, error) {

	diagnostics := diagnostic.Diagnostics{}

	enumInput := &enumeration.EnumerateInput{ResourceTypes: []string{typ}}
	enumerate, err := e.Enumerate(enumInput)
	if err != nil {
		return nil, err
	}
	diagnostics = append(diagnostics, enumerate.Diagnostics...)

	refreshInput := &enumeration.RefreshInput{Resources: enumerate.Resources}
	refresh, err := e.Refresh(refreshInput)
	if err != nil {
		return nil, err
	}
	diagnostics = append(diagnostics, refresh.Diagnostics...)

	return &ListOutput{
		Resources:   refresh.Resources[typ],
		Diagnostics: diagnostics,
	}, nil
}

type sliceAlerter struct {
	lock   sync.Mutex
	alerts alerter.Alerts
}

func newSliceAlerter() *sliceAlerter {
	return &sliceAlerter{
		alerts: alerter.Alerts{},
	}
}

func (d *sliceAlerter) Alerts() alerter.Alerts {
	return d.alerts
}

func (d *sliceAlerter) SendAlert(key string, alert alerter.Alert) {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.alerts[key] = append(d.alerts[key], alert)
}

type typeFilter struct {
	types map[string]struct{}
}

func (u *typeFilter) IsTypeIgnored(ty resource.ResourceType) bool {
	_, ok := u.types[ty.String()]
	return !ok
}

func (u *typeFilter) IsResourceIgnored(res *resource.Resource) bool {
	_, ok := u.types[res.Type]
	return !ok
}

func (u *typeFilter) IsFieldIgnored(res *resource.Resource, path []string) bool {
	return false
}

type dummyCounter struct {
}

func (d *dummyCounter) Inc() {
}

func mapByType(results []*resource.Resource) map[string][]*resource.Resource {
	mapRes := map[string][]*resource.Resource{}
	for _, result := range results {
		mapRes[result.Type] = append(mapRes[result.Type], result)
	}
	return mapRes
}
