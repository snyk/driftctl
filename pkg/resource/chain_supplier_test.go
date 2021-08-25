package resource_test

import (
	"errors"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestChainSupplier_Resources(t *testing.T) {

	fakeTestSupplier := resource.MockSupplier{}
	fakeTestSupplier.On("Resources").Return(
		[]*resource.Resource{
			resource.NewResource("fake-supplier-1_fake-resource-1", "fake_resource_type"),
			resource.NewResource("fake-supplier-1_fake-resource-2", "fake_resource_type"),
		},
		nil,
	).Once()

	anotherFakeTestSupplier := resource.MockSupplier{}
	anotherFakeTestSupplier.On("Resources").Return(
		[]*resource.Resource{
			resource.NewResource("fake-supplier-2_fake-resource-1", "fake_resource_type"),
			resource.NewResource("fake-supplier-2_fake-resource-2", "fake_resource_type"),
		},
		nil,
	).Once()

	chain := resource.NewChainSupplier()
	chain.AddSupplier(&fakeTestSupplier)
	chain.AddSupplier(&anotherFakeTestSupplier)

	res, err := chain.Resources()

	if err != nil {
		t.Fatal(err)
	}

	anotherFakeTestSupplier.AssertExpectations(t)
	fakeTestSupplier.AssertExpectations(t)
	assert.Len(t, res, 4)
}

func TestChainSupplier_Resources_WithError(t *testing.T) {

	fakeTestSupplier := resource.MockSupplier{}
	fakeTestSupplier.
		On("Resources").
		Return([]*resource.Resource{
			resource.NewResource("fake-supplier-1_fake-resource-1", "fake_resource_type"),
			resource.NewResource("fake-supplier-1_fake-resource-2", "fake_resource_type"),
		},
			nil,
		)

	anotherFakeTestSupplier := resource.MockSupplier{}
	anotherFakeTestSupplier.
		On("Resources").
		Return(nil, errors.New("error from another supplier")).
		Once()

	chain := resource.NewChainSupplier()
	chain.AddSupplier(&fakeTestSupplier)
	chain.AddSupplier(&anotherFakeTestSupplier)

	res, err := chain.Resources()

	anotherFakeTestSupplier.AssertExpectations(t)
	assert.Nil(t, res)
	assert.Equal(t, "error from another supplier", err.Error())
}
