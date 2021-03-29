package pkg_test

import (
	"testing"

	"github.com/jmespath/go-jmespath"
	"github.com/r3labs/diff/v2"
	"github.com/zclconf/go-cty/cty"

	"github.com/cloudskiff/driftctl/pkg"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	filter2 "github.com/cloudskiff/driftctl/pkg/filter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/cloudskiff/driftctl/test"
	testresource "github.com/cloudskiff/driftctl/test/resource"
)

type TestCase struct {
	name            string
	stateResources  []resource.Resource
	remoteResources []resource.Resource
	filter          string
	mocks           func(factory resource.ResourceFactory)
	assert          func(result *test.ScanResult, err error)
}

type TestCases []TestCase

func runTest(t *testing.T, cases TestCases) {
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testAlerter := alerter.NewAlerter()

			if c.stateResources == nil {
				c.stateResources = []resource.Resource{}
			}
			stateSupplier := &resource.MockSupplier{}
			stateSupplier.On("Resources").Return(c.stateResources, nil)

			if c.remoteResources == nil {
				c.remoteResources = []resource.Resource{}
			}
			remoteSupplier := &resource.MockSupplier{}
			remoteSupplier.On("Resources").Return(c.remoteResources, nil)

			var filter *jmespath.JMESPath
			if c.filter != "" {
				f, err := filter2.BuildExpression(c.filter)
				if err != nil {
					t.Fatalf("Unable to build filter expression: %s\n%s", c.filter, err)
				}
				filter = f
			}

			resourceFactory := &terraform.MockResourceFactory{}

			if c.mocks != nil {
				c.mocks(resourceFactory)
			}

			driftctl := pkg.NewDriftCTL(remoteSupplier, stateSupplier, filter, testAlerter, resourceFactory)

			analysis, err := driftctl.Run()

			c.assert(test.NewScanResult(t, analysis), err)
		})
	}
}

func TestDriftctlRun_BasicBehavior(t *testing.T) {

	cases := TestCases{
		{
			name: "infrastructure should be in sync",
			stateResources: []resource.Resource{
				testresource.FakeResource{},
			},
			remoteResources: []resource.Resource{
				testresource.FakeResource{},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertInfrastructureIsInSync()
			},
		},
		{
			name: "we should have deleted resource",
			stateResources: []resource.Resource{
				testresource.FakeResource{},
			},
			remoteResources: []resource.Resource{},
			assert: func(result *test.ScanResult, err error) {
				result.AssertDeletedCount(1)
			},
		},
		{
			name:           "we should have unmanaged resource",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
			},
		},
		{
			name: "we should have changes of field update",
			stateResources: []resource.Resource{
				testresource.FakeResource{
					Id:     "fake",
					FooBar: "barfoo",
				},
			},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:     "fake",
					FooBar: "foobar",
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"FooBar"},
						From: "barfoo",
						To:   "foobar",
					},
					Computed: false,
				})
			},
		},
		{
			name: "we should have changes on computed field",
			stateResources: []resource.Resource{
				testresource.FakeResource{
					Id:     "fake",
					BarFoo: "barfoo",
				},
			},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:     "fake",
					BarFoo: "foobar",
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.UPDATE,
						Path: []string{"BarFoo"},
						From: "barfoo",
						To:   "foobar",
					},
					Computed: true,
				})
			},
		},
		{
			name: "we should have changes of deleted field",
			stateResources: []resource.Resource{
				testresource.FakeResource{
					Id: "fake",
					Tags: map[string]string{
						"tag1": "deleted",
					},
				},
			},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id: "fake",
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.DELETE,
						Path: []string{"Tags", "tag1"},
						From: "deleted",
						To:   nil,
					},
					Computed: false,
				})
			},
		},
		{
			name: "we should have changes of added field",
			stateResources: []resource.Resource{
				testresource.FakeResource{
					Id: "fake",
				},
			},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id: "fake",
					Tags: map[string]string{
						"tag1": "added",
					},
				},
			},
			assert: func(result *test.ScanResult, err error) {
				result.AssertManagedCount(1)
				result.AssertResourceHasDrift("fake", "FakeResource", analyser.Change{
					Change: diff.Change{
						Type: diff.CREATE,
						Path: []string{"Tags", "tag1"},
						From: nil,
						To:   "added",
					},
					Computed: false,
				})
			},
		},
	}

	runTest(t, cases)
}

func TestDriftctlRun_BasicFilter(t *testing.T) {
	cases := TestCases{
		{
			name:           "test filtering on Type",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:   "res1",
					Type: "not-filtered",
				},
				testresource.FakeResource{
					Id:   "res2",
					Type: "filtered",
				},
			},
			filter: "Type=='filtered'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
		},
		{
			name:           "test filtering on Id",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:   "res1",
					Type: "not-filtered",
				},
				testresource.FakeResource{
					Id:   "res2",
					Type: "filtered",
				},
			},
			filter: "Id=='res2'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res2", "filtered")
			},
		},
		{
			name:           "test filtering on attribute",
			stateResources: []resource.Resource{},
			remoteResources: []resource.Resource{
				testresource.FakeResource{
					Id:   "res1",
					Type: "filtered",
					CtyVal: func() *cty.Value {
						v := cty.ObjectVal(map[string]cty.Value{
							"test_field": cty.StringVal("value to filter on"),
						})
						return &v
					}(),
				},
				testresource.FakeResource{
					Id:   "res2",
					Type: "not-filtered",
				},
			},
			filter: "Attr.test_field=='value to filter on'",
			assert: func(result *test.ScanResult, err error) {
				result.AssertUnmanagedCount(1)
				result.AssertResourceUnmanaged("res1", "filtered")
			},
		},
	}

	runTest(t, cases)
}
