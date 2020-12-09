package analyser

import (
	"encoding/json"
	"io/ioutil"

	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/resource"

	"github.com/r3labs/diff/v2"

	"testing"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		name       string
		iac        []resource.Resource
		cloud      []resource.Resource
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
				testresource.FakeResource{
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
					testresource.FakeResource{
						Id: "foobar",
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "Test100PercentCoverage",
			iac: []resource.Resource{
				testresource.FakeResource{
					Id: "foobar",
				},
			},
			cloud: []resource.Resource{
				testresource.FakeResource{
					Id: "foobar",
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					testresource.FakeResource{
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
				testresource.FakeResource{
					Id: "foobar",
				},
			},
			expected: Analysis{
				summary: Summary{
					TotalResources: 1,
					TotalUnmanaged: 1,
				},
				unmanaged: []resource.Resource{
					testresource.FakeResource{
						Id: "foobar",
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestDiff",
			iac: []resource.Resource{
				testresource.FakeResource{
					Id:     "foobar",
					FooBar: "foobar",
					BarFoo: "barfoo",
				},
			},
			cloud: []resource.Resource{
				testresource.FakeResource{
					Id:     "foobar",
					FooBar: "barfoo",
					BarFoo: "foobar",
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					testresource.FakeResource{
						Id:     "foobar",
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
						Res: testresource.FakeResource{
							Id:     "foobar",
							FooBar: "foobar",
							BarFoo: "barfoo",
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
						},
					},
				},
			},
			hasDrifted: true,
		},
	}

	analyzer := NewAnalyzer()

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			result, err := analyzer.Analyze(c.cloud, c.iac)

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

			summaryChanges, err := diff.Diff(result.Summary(), c.expected.Summary())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(summaryChanges) > 0 {
				for _, change := range summaryChanges {
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
		testresource.FakeResource{
			Id:   "AKIA5QYBVVD25KFXJHYJ",
			Type: "aws_iam_access_key",
		}, testresource.FakeResource{
			Id:   "driftctl2",
			Type: "aws_managed_resource",
		},
	)
	analysis.AddUnmanaged(
		testresource.FakeResource{
			Id:   "driftctl",
			Type: "aws_s3_bucket_policy",
		}, testresource.FakeResource{
			Id:   "driftctl",
			Type: "aws_s3_bucket_notification",
		},
	)
	analysis.AddDeleted(
		testresource.FakeResource{
			Id:     "test-driftctl2",
			Type:   "aws_iam_user",
			FooBar: "test",
		},
		testresource.FakeResource{
			Id:   "AKIA5QYBVVD2Y6PBAAPY",
			Type: "aws_iam_access_key",
		},
	)
	analysis.AddDifference(Difference{
		Res: testresource.FakeResource{
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
