package remote

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
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
