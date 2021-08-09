package resource_test

import (
	"errors"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestChainSupplier_Resources(t *testing.T) {

	assert := assert.New(t)

	fakeTestSupplier := resource.MockSupplier{}
	fakeTestSupplier.On("Resources").Return(
		[]*resource.Resource{
			&resource.Resource{
				Id: "fake-supplier-1_fake-resource-1",
			},
			&resource.Resource{
				Id: "fake-supplier-1_fake-resource-2",
			},
		},
		nil,
	).Once()

	anotherFakeTestSupplier := resource.MockSupplier{}
	anotherFakeTestSupplier.On("Resources").Return(
		[]*resource.Resource{
			&resource.Resource{
				Id: "fake-supplier-2_fake-resource-1",
			},
			&resource.Resource{
				Id: "fake-supplier-2_fake-resource-2",
			},
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
	assert.Len(res, 4)
}

func TestChainSupplier_Resources_WithError(t *testing.T) {

	assert := assert.New(t)

	fakeTestSupplier := resource.MockSupplier{}
	fakeTestSupplier.
		On("Resources").
		Return([]*resource.Resource{
			&resource.Resource{
				Id: "fake-supplier-1_fake-resource-1",
			},
			&resource.Resource{
				Id: "fake-supplier-1_fake-resource-2",
			},
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
	assert.Nil(res)
	assert.Equal("error from another supplier", err.Error())
}
