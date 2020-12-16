package analyser

import (
	"encoding/json"
	"io/ioutil"

	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/r3labs/diff/v2"

	"testing"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		name         string
		iac          []resource.Resource
		ignoredRes   []resource.Resource
		cloud        []resource.Resource
		ignoredDrift []struct {
			res  resource.Resource
			path []string
		}
		alerts     alerter.Alerts
		expected   Analysis
		hasDrifted bool
	}{
		{
			name:     "TestNilValues", // Cover division by zero case
			iac:      nil,
			cloud:    nil,
			expected: Analysis{},
		},
		{
			name:     "TestNothingToCompare", // Cover division by zero case
			iac:      []resource.Resource{},
			cloud:    []resource.Resource{},
			expected: Analysis{},
		},
		{
			name: "TestIgnoreFromCoverageIacNotInCloud",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			cloud: []resource.Resource{},
			expected: Analysis{
				summary: Summary{
					TotalResources: 1,
					TotalDeleted:   1,
				},
				deleted: []resource.Resource{
					&testresource.FakeResource{
						Id: "foobar",
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestResourceIgnoredDeleted",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			ignoredRes: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			cloud: []resource.Resource{},
			expected: Analysis{
				summary: Summary{
					TotalResources: 0,
					TotalDeleted:   0,
				},
			},
			hasDrifted: false,
		},
		{
			name: "Test100PercentCoverage with ignore",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
				&testresource.FakeResource{
					Id: "foobar2",
				},
			},
			ignoredRes: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar2",
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
				&testresource.FakeResource{
					Id: "foobar2",
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id: "foobar",
					},
				},
				summary: Summary{
					TotalManaged:   1,
					TotalResources: 1,
				},
			},
		},
		{
			name: "Test100PercentCoverage",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id: "foobar",
					},
				},
				summary: Summary{
					TotalManaged:   1,
					TotalResources: 1,
				},
			},
		},
		{
			name: "TestUnmanagedResource",
			iac:  []resource.Resource{},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
				},
			},
			expected: Analysis{
				summary: Summary{
					TotalResources: 1,
					TotalUnmanaged: 1,
				},
				unmanaged: []resource.Resource{
					&testresource.FakeResource{
						Id: "foobar",
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestDiff",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"baz", "bar"},
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"bar", "baz"},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id:     "foobar",
						FooBar: "foobar",
						BarFoo: "barfoo",
						Struct: struct {
							Baz string `computed:"true"`
							Bar string
						}{"baz", "bar"},
					},
				},
				summary: Summary{
					TotalResources: 1,
					TotalDrifted:   1,
					TotalManaged:   1,
				},
				differences: []Difference{
					{
						Res: &testresource.FakeResource{
							Id:     "foobar",
							FooBar: "foobar",
							BarFoo: "barfoo",
							Struct: struct {
								Baz string `computed:"true"`
								Bar string
							}{"baz", "bar"},
						},
						Changelog: diff.Changelog{
							diff.Change{
								Type: "update",
								From: "foobar",
								To:   "barfoo",
								Path: []string{
									"FooBar",
								},
							},
							diff.Change{
								Type: "update",
								From: "barfoo",
								To:   "foobar",
								Path: []string{
									"BarFoo",
								},
							},
							diff.Change{
								Type: "update",
								From: "baz",
								To:   "bar",
								Path: []string{
									"Struct",
									"Baz",
								},
							},
							diff.Change{
								Type: "update",
								From: "bar",
								To:   "baz",
								Path: []string{
									"Struct",
									"Bar",
								},
							},
						},
					},
				},
				alerts: alerter.Alerts{
					"FakeResource.foobar": {
						{
							Message: "BarFoo is a computed field",
						},
						{
							Message: "Struct.Baz is a computed field",
						},
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestDiff with partial ignore",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					Type:   "fakeres",
					FooBar: "foobar",
					BarFoo: "barfoo",
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					Type:   "fakeres",
					FooBar: "barfoo",
					BarFoo: "foobar",
				},
			},
			ignoredDrift: []struct {
				res  resource.Resource
				path []string
			}{
				{
					res: &testresource.FakeResource{
						Id:     "foobar",
						Type:   "fakeres",
						FooBar: "foobar",
						BarFoo: "barfoo",
					},
					path: []string{"FooBar"},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id:     "foobar",
						Type:   "fakeres",
						FooBar: "foobar",
						BarFoo: "barfoo",
					},
				},
				summary: Summary{
					TotalResources: 1,
					TotalDrifted:   1,
					TotalManaged:   1,
				},
				differences: []Difference{
					{
						Res: &testresource.FakeResource{
							Id:     "foobar",
							Type:   "fakeres",
							FooBar: "foobar",
							BarFoo: "barfoo",
						},
						Changelog: diff.Changelog{
							diff.Change{
								Type: "update",
								From: "barfoo",
								To:   "foobar",
								Path: []string{
									"BarFoo",
								},
							},
						},
					},
				},
				alerts: alerter.Alerts{
					"fakeres.foobar": {
						{
							Message: "BarFoo is a computed field",
						},
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestDiff with full ignore",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					FooBar: "foobar",
					BarFoo: "barfoo",
				},
			},
			ignoredRes: []resource.Resource{
				&testresource.FakeResource{
					Id: "should_be_ignored",
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					FooBar: "barfoo",
					BarFoo: "foobar",
				},
				&testresource.FakeResource{
					Id: "should_be_ignored",
				},
			},
			ignoredDrift: []struct {
				res  resource.Resource
				path []string
			}{
				{
					res: &testresource.FakeResource{
						Id:     "foobar",
						FooBar: "foobar",
						BarFoo: "barfoo",
					},
					path: []string{"FooBar"},
				},
				{
					res: &testresource.FakeResource{
						Id:     "foobar",
						FooBar: "foobar",
						BarFoo: "barfoo",
					},
					path: []string{"BarFoo"},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id:     "foobar",
						FooBar: "foobar",
						BarFoo: "barfoo",
					},
				},
				summary: Summary{
					TotalResources: 1,
					TotalDrifted:   0,
					TotalManaged:   1,
				},
			},
			hasDrifted: false,
		},
		{
			name: "TestDiffWithAlertFiltering",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					Type:   "fakeres",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "barfoo",
					Type:   "fakeres",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "foobaz",
					Type:   "other",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "resource",
					Type:   "other",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"baz", "bar"},
					StructSlice: []struct {
						String string   `computed:"true"`
						Array  []string `computed:"true"`
					}{
						{"one", []string{"foo"}},
					},
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					Type:   "fakeres",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"bar", "baz"},
				},
				&testresource.FakeResource{
					Id:     "barfoo",
					Type:   "fakeres",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"bar", "baz"},
				},
				&testresource.FakeResource{
					Id:     "foobaz",
					Type:   "other",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"bar", "baz"},
				},
				&testresource.FakeResource{
					Id:     "resource",
					Type:   "other",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `computed:"true"`
						Bar string
					}{"bar", "baz"},
					StructSlice: []struct {
						String string   `computed:"true"`
						Array  []string `computed:"true"`
					}{
						{"two", []string{"oof"}},
					},
				},
			},
			alerts: alerter.Alerts{
				"fakeres": {
					{
						Message:              "Should be ignored",
						ShouldIgnoreResource: true,
					},
				},
				"other.foobaz": {
					{
						Message:              "Should be ignored",
						ShouldIgnoreResource: true,
					},
				},
				"other.resource": {
					{
						Message: "Should not be ignored",
					},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id:     "resource",
						Type:   "other",
						FooBar: "foobar",
						BarFoo: "barfoo",
						Struct: struct {
							Baz string `computed:"true"`
							Bar string
						}{"baz", "bar"},
						StructSlice: []struct {
							String string   `computed:"true"`
							Array  []string `computed:"true"`
						}{
							{"one", []string{"foo"}},
						},
					},
				},
				summary: Summary{
					TotalResources: 1,
					TotalDrifted:   1,
					TotalManaged:   1,
				},
				differences: []Difference{
					{
						Res: &testresource.FakeResource{
							Id:     "resource",
							Type:   "other",
							FooBar: "foobar",
							BarFoo: "barfoo",
							Struct: struct {
								Baz string `computed:"true"`
								Bar string
							}{"baz", "bar"},
							StructSlice: []struct {
								String string   `computed:"true"`
								Array  []string `computed:"true"`
							}{
								{"one", []string{"foo"}},
							},
						},
						Changelog: diff.Changelog{
							diff.Change{
								Type: "update",
								From: "foobar",
								To:   "barfoo",
								Path: []string{
									"FooBar",
								},
							},
							diff.Change{
								Type: "update",
								From: "barfoo",
								To:   "foobar",
								Path: []string{
									"BarFoo",
								},
							},
							diff.Change{
								Type: "update",
								From: "baz",
								To:   "bar",
								Path: []string{
									"Struct",
									"Baz",
								},
							},
							diff.Change{
								Type: "update",
								From: "bar",
								To:   "baz",
								Path: []string{
									"Struct",
									"Bar",
								},
							},
							diff.Change{
								Type: "update",
								From: "foo",
								To:   "oof",
								Path: []string{
									"StructSlice",
									"0",
									"Array",
									"0",
								},
							},
							diff.Change{
								Type: "update",
								From: "one",
								To:   "two",
								Path: []string{
									"StructSlice",
									"0",
									"String",
								},
							},
						},
					},
				},
				alerts: alerter.Alerts{
					"fakeres": {
						{
							Message:              "Should be ignored",
							ShouldIgnoreResource: true,
						},
					},
					"other.foobaz": {
						{
							Message:              "Should be ignored",
							ShouldIgnoreResource: true,
						},
					},
					"other.resource": {
						{
							Message: "Should not be ignored",
						},
						{
							Message: "BarFoo is a computed field",
						},
						{
							Message: "Struct.Baz is a computed field",
						},
						{
							Message: "StructSlice.0.String is a computed field",
						},
						{
							Message: "StructSlice.0.Array.0 is a computed field",
						},
					},
				},
			},
			hasDrifted: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			filter := &mocks.Filter{}
			for _, ignored := range c.ignoredRes {
				filter.On("IsResourceIgnored", ignored).Return(true)
			}
			filter.On("IsResourceIgnored", mock.Anything).Return(false)

			for _, s := range c.ignoredDrift {
				filter.On("IsFieldIgnored", s.res, s.path).Return(true)
			}
			filter.On("IsFieldIgnored", mock.Anything, mock.Anything).Return(false)

			alerter := alerter.NewAlerter()
			if c.alerts != nil {
				alerter.SetAlerts(c.alerts)
			}

			analyzer := NewAnalyzer(alerter)
			result, err := analyzer.Analyze(c.cloud, c.iac, filter)

			if err != nil {
				t.Error(err)
				return
			}

			if result.IsSync() == c.hasDrifted {
				t.Errorf("Drifted state does not match, got %t expected %t", result.IsSync(), !c.hasDrifted)
			}

			managedChanges, err := diff.Diff(result.Managed(), c.expected.Managed())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(managedChanges) > 0 {
				for _, change := range managedChanges {
					t.Errorf("%+v", change)
				}
			}

			unmanagedChanges, err := diff.Diff(result.Unmanaged(), c.expected.Unmanaged())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(unmanagedChanges) > 0 {
				for _, change := range unmanagedChanges {
					t.Errorf("%+v", change)
				}
			}

			deletedChanges, err := diff.Diff(result.Deleted(), c.expected.Deleted())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(deletedChanges) > 0 {
				for _, change := range deletedChanges {
					t.Errorf("%+v", change)
				}
			}

			diffChanges, err := diff.Diff(result.Differences(), c.expected.Differences())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(diffChanges) > 0 {
				for _, change := range diffChanges {
					t.Errorf("%+v", change)
				}
			}

			summaryChanges, err := diff.Diff(c.expected.Summary(), result.Summary())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(summaryChanges) > 0 {
				for _, change := range summaryChanges {
					t.Errorf("%+v", change)
				}
			}

			alertsChanges, err := diff.Diff(result.Alerts(), c.expected.Alerts())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(alertsChanges) > 0 {
				for _, change := range alertsChanges {
					t.Errorf("%+v", change)
				}
			}
		})
	}
}

func TestAnalysis_MarshalJSON(t *testing.T) {
	goldenFile := "./testdata/output.json"
	analysis := Analysis{}
	analysis.AddManaged(
		&testresource.FakeResource{
			Id:   "AKIA5QYBVVD25KFXJHYJ",
			Type: "aws_iam_access_key",
		}, &testresource.FakeResource{
			Id:   "driftctl2",
			Type: "aws_managed_resource",
		},
	)
	analysis.AddUnmanaged(
		&testresource.FakeResource{
			Id:   "driftctl",
			Type: "aws_s3_bucket_policy",
		}, &testresource.FakeResource{
			Id:   "driftctl",
			Type: "aws_s3_bucket_notification",
		},
	)
	analysis.AddDeleted(
		&testresource.FakeResource{
			Id:     "test-driftctl2",
			Type:   "aws_iam_user",
			FooBar: "test",
		},
		&testresource.FakeResource{
			Id:   "AKIA5QYBVVD2Y6PBAAPY",
			Type: "aws_iam_access_key",
		},
	)
	analysis.AddDifference(Difference{
		Res: &testresource.FakeResource{
			Id:   "AKIA5QYBVVD25KFXJHYJ",
			Type: "aws_iam_access_key",
		},
		Changelog: []diff.Change{
			{
				Type: "update",
				Path: []string{"Status"},
				From: "Active",
				To:   "Inactive",
			},
		},
	})
	analysis.AddAlerts(alerter.Alerts{
		"aws_iam_access_key": {
			{
				Message: "This is an alert",
			},
		},
	})

	got, err := json.MarshalIndent(analysis, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	if *goldenfile.Update == "TestAnalysis_MarshalJSON" {
		if err := ioutil.WriteFile(goldenFile, got, 0600); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := ioutil.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, err)
	assert.Equal(t, string(expected), string(got))
}

func TestAnalysis_UnmarshalJSON(t *testing.T) {
	expected := Analysis{
		summary: Summary{
			TotalResources: 6,
			TotalDrifted:   1,
			TotalUnmanaged: 2,
			TotalDeleted:   2,
			TotalManaged:   2,
		},
		managed: []resource.Resource{
			resource.SerializedResource{
				Id:   "AKIA5QYBVVD25KFXJHYJ",
				Type: "aws_iam_access_key",
			},
			resource.SerializedResource{
				Id:   "test-managed",
				Type: "aws_iam_user",
			},
		},
		unmanaged: []resource.Resource{
			resource.SerializedResource{
				Id:   "driftctl",
				Type: "aws_s3_bucket_policy",
			},
			resource.SerializedResource{
				Id:   "driftctl",
				Type: "aws_s3_bucket_notification",
			},
		},
		deleted: []resource.Resource{
			resource.SerializedResource{
				Id:   "test-driftctl2",
				Type: "aws_iam_user",
			},
			resource.SerializedResource{
				Id:   "AKIA5QYBVVD2Y6PBAAPY",
				Type: "aws_iam_access_key",
			},
		},
		differences: []Difference{
			{
				Res: resource.SerializedResource{
					Id:   "AKIA5QYBVVD25KFXJHYJ",
					Type: "aws_iam_access_key",
				},
				Changelog: []diff.Change{
					{
						Type: "update",
						Path: []string{"Status"},
						From: "Active",
						To:   "Inactive",
					},
				},
			},
		},
		alerts: alerter.Alerts{
			"aws_iam_access_key": {
				{
					Message: "This is an alert",
				},
			},
		},
	}

	got := Analysis{}
	input, err := ioutil.ReadFile("./testdata/input.json")
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(input, &got)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, got)
	assert.Equal(t, 33, got.Coverage())
	assert.Equal(t, 2, got.Summary().TotalUnmanaged)
	assert.Equal(t, 2, got.Summary().TotalManaged)
	assert.Equal(t, 2, got.Summary().TotalDeleted)
	assert.Equal(t, 6, got.Summary().TotalResources)
	assert.Equal(t, 1, got.Summary().TotalDrifted)
}
