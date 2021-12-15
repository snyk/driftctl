package remote

import (
	"testing"

	"github.com/snyk/driftctl/pkg/alerter"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestScannerShouldIgnoreType(t *testing.T) {

	// Initialize mocks
	alerter := alerter.NewAlerter()
	fakeEnumerator := &common.MockEnumerator{}
	fakeEnumerator.On("SupportedType").Return(resource.ResourceType("FakeType"))
	fakeEnumerator.AssertNotCalled(t, "Enumerate")

	remoteLibrary := common.NewRemoteLibrary()
	remoteLibrary.AddEnumerator(fakeEnumerator)

	testFilter := &filter.MockFilter{}
	testFilter.On("IsTypeIgnored", resource.ResourceType("FakeType")).Return(true)

	s := NewScanner(remoteLibrary, alerter, ScannerOptions{}, testFilter)
	_, err := s.Resources()
	assert.Nil(t, err)
	fakeEnumerator.AssertExpectations(t)
}

func TestScannerShouldReadManagedOnly(t *testing.T) {

	Resources := []*resource.Resource{
		{
			Id:    "test-1",
			Type:  "FakeType",
			Attrs: &resource.Attributes{},
		},
		{
			Id:    "test-2",
			Type:  "FakeType",
			Attrs: &resource.Attributes{},
		},
	}

	// Initialize mocks
	fakeEnumerator := &common.MockEnumerator{}
	fakeEnumerator.On("SupportedType").Return(resource.ResourceType("FakeType"))
	fakeEnumerator.On("Enumerate").Return(Resources, nil)

	fakeDetailsFetcher := &common.MockDetailsFetcher{}
	fakeDetailsFetcher.On("ReadDetails", Resources[1]).Return(Resources[1], nil)

	remoteLibrary := common.NewRemoteLibrary()
	remoteLibrary.AddEnumerator(fakeEnumerator)
	remoteLibrary.AddDetailsFetcher("FakeType", fakeDetailsFetcher)

	testFilter := &filter.MockFilter{}
	testFilter.On("IsTypeIgnored", resource.ResourceType("FakeType")).Return(false)

	s := NewScanner(remoteLibrary, alerter.NewAlerter(), ScannerOptions{Deep: true}, testFilter)
	remoteResources, err := s.EnumerateResources()
	assert.Nil(t, err)

	remoteResources, err = s.ReadResources(remoteResources[1:])
	assert.Nil(t, err)

	assert.Equal(t, Resources[1:], remoteResources)

	fakeEnumerator.AssertExpectations(t)
	fakeDetailsFetcher.AssertExpectations(t)
	testFilter.AssertExpectations(t)
}
