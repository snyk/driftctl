package analyser

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/cloudskiff/driftctl/mocks"

	"github.com/stretchr/testify/assert"

	testresource "github.com/cloudskiff/driftctl/test/resource"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/r3labs/diff/v2"
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
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
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
							Baz string `cty:"baz" computed:"true"`
							Bar string `cty:"bar"`
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
								Baz string `cty:"baz" computed:"true"`
								Bar string `cty:"bar"`
							}{"baz", "bar"},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "foobar",
									To:   "barfoo",
									Path: []string{
										"FooBar",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "barfoo",
									To:   "foobar",
									Path: []string{
										"BarFoo",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "baz",
									To:   "bar",
									Path: []string{
										"Struct",
										"Baz",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
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
				},
				alerts: alerter.Alerts{
					"": {
						NewComputedDiffAlert(),
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
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "barfoo",
									To:   "foobar",
									Path: []string{
										"BarFoo",
									},
								},
								Computed: true,
							},
						},
					},
				},
				alerts: alerter.Alerts{
					"": {
						NewComputedDiffAlert(),
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
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "barfoo",
					Type:   "fakeres",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "foobaz",
					Type:   "other",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "resource",
					Type:   "other",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
					StructSlice: []struct {
						String string   `cty:"string" computed:"true"`
						Array  []string `cty:"array" computed:"true"`
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
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"bar", "baz"},
				},
				&testresource.FakeResource{
					Id:     "barfoo",
					Type:   "fakeres",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"bar", "baz"},
				},
				&testresource.FakeResource{
					Id:     "foobaz",
					Type:   "other",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"bar", "baz"},
				},
				&testresource.FakeResource{
					Id:     "resource",
					Type:   "other",
					FooBar: "barfoo",
					BarFoo: "foobar",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"bar", "baz"},
					StructSlice: []struct {
						String string   `cty:"string" computed:"true"`
						Array  []string `cty:"array" computed:"true"`
					}{
						{"two", []string{"oof"}},
					},
				},
			},
			alerts: alerter.Alerts{
				"fakeres": {
					&alerter.FakeAlert{Msg: "Should be ignored", IgnoreResource: true},
				},
				"other.foobaz": {
					&alerter.FakeAlert{Msg: "Should be ignored", IgnoreResource: true},
				},
				"other.resource": {
					&alerter.FakeAlert{Msg: "Should not be ignored"},
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
							Baz string `cty:"baz" computed:"true"`
							Bar string `cty:"bar"`
						}{"baz", "bar"},
						StructSlice: []struct {
							String string   `cty:"string" computed:"true"`
							Array  []string `cty:"array" computed:"true"`
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
								Baz string `cty:"baz" computed:"true"`
								Bar string `cty:"bar"`
							}{"baz", "bar"},
							StructSlice: []struct {
								String string   `cty:"string" computed:"true"`
								Array  []string `cty:"array" computed:"true"`
							}{
								{"one", []string{"foo"}},
							},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "foobar",
									To:   "barfoo",
									Path: []string{
										"FooBar",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "barfoo",
									To:   "foobar",
									Path: []string{
										"BarFoo",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "baz",
									To:   "bar",
									Path: []string{
										"Struct",
										"Baz",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "bar",
									To:   "baz",
									Path: []string{
										"Struct",
										"Bar",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "one",
									To:   "two",
									Path: []string{
										"StructSlice",
										"0",
										"String",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
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
								Computed: true,
							},
						},
					},
				},
				alerts: alerter.Alerts{
					"fakeres": {
						&alerter.FakeAlert{Msg: "Should be ignored", IgnoreResource: true},
					},
					"other.foobaz": {
						&alerter.FakeAlert{Msg: "Should be ignored", IgnoreResource: true},
					},
					"other.resource": {
						&alerter.FakeAlert{Msg: "Should not be ignored"},
					},
					"": {
						NewComputedDiffAlert(),
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestDiff with computed field send 1 alert",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					Type:   "fakeres",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
				},
				&testresource.FakeResource{
					Id:     "resource",
					Type:   "other",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"baz", "bar"},
					StructSlice: []struct {
						String string   `cty:"string" computed:"true"`
						Array  []string `cty:"array" computed:"true"`
					}{
						{"one", []string{"foo"}},
					},
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:     "foobar",
					Type:   "fakeres",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"bazdiff", "bardiff"},
				},
				&testresource.FakeResource{
					Id:     "resource",
					Type:   "other",
					FooBar: "foobar",
					BarFoo: "barfoo",
					Struct: struct {
						Baz string `cty:"baz" computed:"true"`
						Bar string `cty:"bar"`
					}{"bazdiff", "bar"},
					StructSlice: []struct {
						String string   `cty:"string" computed:"true"`
						Array  []string `cty:"array" computed:"true"`
					}{
						{"onediff", []string{"foo", "diff"}},
					},
				},
			},
			alerts: alerter.Alerts{},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id:     "foobar",
						Type:   "fakeres",
						FooBar: "foobar",
						BarFoo: "barfoo",
						Struct: struct {
							Baz string `cty:"baz" computed:"true"`
							Bar string `cty:"bar"`
						}{"baz", "bar"},
					},
					&testresource.FakeResource{
						Id:     "resource",
						Type:   "other",
						FooBar: "foobar",
						BarFoo: "barfoo",
						Struct: struct {
							Baz string `cty:"baz" computed:"true"`
							Bar string `cty:"bar"`
						}{"baz", "bar"},
						StructSlice: []struct {
							String string   `cty:"string" computed:"true"`
							Array  []string `cty:"array" computed:"true"`
						}{
							{"one", []string{"foo"}},
						},
					},
				},
				summary: Summary{
					TotalResources: 2,
					TotalDrifted:   2,
					TotalManaged:   2,
				},
				differences: []Difference{
					{
						Res: &testresource.FakeResource{
							Id:     "foobar",
							Type:   "fakeres",
							FooBar: "foobar",
							BarFoo: "barfoo",
							Struct: struct {
								Baz string `cty:"baz" computed:"true"`
								Bar string `cty:"bar"`
							}{"baz", "bar"},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "baz",
									To:   "bazdiff",
									Path: []string{
										"Struct",
										"Baz",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "bar",
									To:   "bardiff",
									Path: []string{
										"Struct",
										"Bar",
									},
								},
								Computed: false,
							},
						},
					},
					{
						Res: &testresource.FakeResource{
							Id:     "resource",
							Type:   "other",
							FooBar: "foobar",
							BarFoo: "barfoo",
							Struct: struct {
								Baz string `cty:"baz" computed:"true"`
								Bar string `cty:"bar"`
							}{"baz", "bar"},
							StructSlice: []struct {
								String string   `cty:"string" computed:"true"`
								Array  []string `cty:"array" computed:"true"`
							}{
								{"one", []string{"foo"}},
							},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "baz",
									To:   "bazdiff",
									Path: []string{
										"Struct",
										"Baz",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "one",
									To:   "onediff",
									Path: []string{
										"StructSlice",
										"0",
										"String",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "create",
									From: nil,
									To:   "diff",
									Path: []string{
										"StructSlice",
										"0",
										"Array",
										"1",
									},
								},
								Computed: true,
							},
						},
					},
				},
				alerts: alerter.Alerts{
					"": {
						NewComputedDiffAlert(),
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "Test alert on unmanaged security group rules",
			iac: []resource.Resource{
				&aws.AwsSecurityGroup{
					Id: "managed security group",
				},
			},
			cloud: []resource.Resource{
				&aws.AwsSecurityGroup{
					Id: "managed security group",
				},
				&aws.AwsSecurityGroupRule{
					Id: "unmanaged rule",
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&aws.AwsSecurityGroup{
						Id: "managed security group",
					},
				},
				unmanaged: []resource.Resource{
					&aws.AwsSecurityGroupRule{
						Id: "unmanaged rule",
					},
				},
				summary: Summary{
					TotalResources: 2,
					TotalManaged:   1,
					TotalUnmanaged: 1,
				},
				alerts: alerter.Alerts{
					"": {
						newUnmanagedSecurityGroupRulesAlert(),
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "Test sorted unmanaged & deleted resources",
			iac: []resource.Resource{
				&testresource.FakeResource{
					Id:   "deleted resource 22",
					Type: "aws_s3_bucket",
				},
				&testresource.FakeResource{
					Id:   "deleted resource 20",
					Type: "aws_ebs_volume",
				},
				&testresource.FakeResource{
					Id:   "deleted resource 20",
					Type: "aws_s3_bucket",
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:   "unmanaged resource 12",
					Type: "aws_s3_bucket",
				},
				&testresource.FakeResource{
					Id:   "unmanaged resource 10",
					Type: "aws_s3_bucket",
				},
				&testresource.FakeResource{
					Id:   "unmanaged resource 11",
					Type: "aws_ebs_volume",
				},
			},
			expected: Analysis{
				managed: []resource.Resource{},
				unmanaged: []resource.Resource{
					&testresource.FakeResource{
						Id:   "unmanaged resource 11",
						Type: "aws_ebs_volume",
					},
					&testresource.FakeResource{
						Id:   "unmanaged resource 10",
						Type: "aws_s3_bucket",
					},
					&testresource.FakeResource{
						Id:   "unmanaged resource 12",
						Type: "aws_s3_bucket",
					},
				},
				deleted: []resource.Resource{
					&testresource.FakeResource{
						Id:   "deleted resource 20",
						Type: "aws_ebs_volume",
					},
					&testresource.FakeResource{
						Id:   "deleted resource 20",
						Type: "aws_s3_bucket",
					},
					&testresource.FakeResource{
						Id:   "deleted resource 22",
						Type: "aws_s3_bucket",
					},
				},
				summary: Summary{
					TotalResources: 6,
					TotalManaged:   0,
					TotalUnmanaged: 3,
					TotalDeleted:   3,
				},
				alerts: alerter.Alerts{},
			},
			hasDrifted: true,
		},
	}

	differ, err := diff.NewDiffer(diff.SliceOrdering(true))
	if err != nil {
		t.Fatalf("Error creating new differ: %e", err)
	}

	differ, err := diff.NewDiffer(diff.SliceOrdering(true))
	if err != nil {
		t.Fatalf("Error creating new differ: %e", err)
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

			al := alerter.NewAlerter()
			if c.alerts != nil {
				al.SetAlerts(c.alerts)
			}

			analyzer := NewAnalyzer(al)
			result, err := analyzer.Analyze(c.cloud, c.iac, filter)

			if err != nil {
				t.Error(err)
				return
			}

			if result.IsSync() == c.hasDrifted {
				t.Errorf("Drifted state does not match, got %t expected %t", result.IsSync(), !c.hasDrifted)
			}

			managedChanges, err := differ.Diff(result.Managed(), c.expected.Managed())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(managedChanges) > 0 {
				for _, change := range managedChanges {
					t.Errorf("%+v", change)
				}
			}

			unmanagedChanges, err := differ.Diff(result.Unmanaged(), c.expected.Unmanaged())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(unmanagedChanges) > 0 {
				for _, change := range unmanagedChanges {
					t.Errorf("%+v", change)
				}
			}

			deletedChanges, err := differ.Diff(result.Deleted(), c.expected.Deleted())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(deletedChanges) > 0 {
				for _, change := range deletedChanges {
					t.Errorf("%+v", change)
				}
			}

			diffChanges, err := differ.Diff(result.Differences(), c.expected.Differences())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(diffChanges) > 0 {
				for _, change := range diffChanges {
					t.Errorf("%+v", change)
				}
			}

			summaryChanges, err := differ.Diff(c.expected.Summary(), result.Summary())
			if err != nil {
				t.Fatalf("Unable to compare %+v", err)
			}
			if len(summaryChanges) > 0 {
				for _, change := range summaryChanges {
					t.Errorf("%+v", change)
				}
			}

			alertsChanges, err := differ.Diff(result.Alerts(), c.expected.Alerts())
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
		Changelog: []Change{
			{
				Change: diff.Change{
					Type: "update",
					Path: []string{"Status"},
					From: "Active",
					To:   "Inactive",
				},
			},
		},
	})
	analysis.SetAlerts(alerter.Alerts{
		"aws_iam_access_key": {
			&alerter.FakeAlert{Msg: "This is an alert"},
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
				Changelog: []Change{
					{
						Change: diff.Change{
							Type: "update",
							Path: []string{"Status"},
							From: "Active",
							To:   "Inactive",
						},
					},
				},
			},
		},
		alerts: alerter.Alerts{
			"aws_iam_access_key": {
				&alerter.SerializedAlert{
					Msg: "This is an alert",
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
	assert.Len(t, got.alerts, 1)
	assert.Equal(t, got.alerts["aws_iam_access_key"][0].Message(), "This is an alert")
}
