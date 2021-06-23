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
				&resource.AbstractResource{
					Id:   "foobar",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"architecture": "foobar",
						"arn":          "barfoo",
						"ebs_block_device": []map[string]interface{}{
							{
								"volume_type": "bar",
								"volume_size": 0,
							},
						},
					},
				},
			},
			cloud: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foobar",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"architecture": "barfoo",
						"arn":          "foobar",
						"ebs_block_device": []map[string]interface{}{
							{
								"volume_type": "baz",
								"volume_size": 1,
							},
						},
					},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&resource.AbstractResource{
						Id:   "foobar",
						Type: aws.AwsAmiResourceType,
						Attrs: &resource.Attributes{
							"architecture": "foobar",
							"arn":          "barfoo",
							"ebs_block_device": []map[string]interface{}{
								{
									"volume_type": "bar",
									"volume_size": 0,
								},
							},
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
						Res: &resource.AbstractResource{
							Id:   "foobar",
							Type: aws.AwsAmiResourceType,
							Attrs: &resource.Attributes{
								"architecture": "foobar",
								"arn":          "barfoo",
								"ebs_block_device": []map[string]interface{}{
									{
										"volume_type": "bar",
										"volume_size": 0,
									},
								},
							},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "foobar",
									To:   "barfoo",
									Path: []string{
										"architecture",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "barfoo",
									To:   "foobar",
									Path: []string{
										"arn",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: 0,
									To:   1,
									Path: []string{
										"ebs_block_device",
										"0",
										"volume_size",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "bar",
									To:   "baz",
									Path: []string{
										"ebs_block_device",
										"0",
										"volume_type",
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
				&resource.AbstractResource{
					Id:   "foobar",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"architecture": "foobar",
						"arn":          "barfoo",
					},
				},
			},
			cloud: []resource.Resource{
				&resource.AbstractResource{
					Id:   "foobar",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"architecture": "barfoo",
						"arn":          "foobar",
					},
				},
			},
			ignoredDrift: []struct {
				res  resource.Resource
				path []string
			}{
				{
					res: &resource.AbstractResource{
						Id:   "foobar",
						Type: aws.AwsAmiResourceType,
						Attrs: &resource.Attributes{
							"architecture": "foobar",
							"arn":          "barfoo",
						},
					},
					path: []string{"architecture"},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&resource.AbstractResource{
						Id:   "foobar",
						Type: aws.AwsAmiResourceType,
						Attrs: &resource.Attributes{
							"architecture": "foobar",
							"arn":          "barfoo",
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
						Res: &resource.AbstractResource{
							Id:   "foobar",
							Type: aws.AwsAmiResourceType,
							Attrs: &resource.Attributes{
								"architecture": "foobar",
								"arn":          "barfoo",
							},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "barfoo",
									To:   "foobar",
									Path: []string{
										"arn",
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
					Id: "foobar",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
						"barfoo": "barfoo",
					},
				},
			},
			ignoredRes: []resource.Resource{
				&testresource.FakeResource{
					Id:    "should_be_ignored",
					Attrs: &resource.Attributes{},
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id: "foobar",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
						"barfoo": "foobar",
					},
				},
				&testresource.FakeResource{
					Id:    "should_be_ignored",
					Attrs: &resource.Attributes{},
				},
			},
			ignoredDrift: []struct {
				res  resource.Resource
				path []string
			}{
				{
					res: &testresource.FakeResource{
						Id: "foobar",
						Attrs: &resource.Attributes{
							"foobar": "foobar",
							"barfoo": "barfoo",
						},
					},
					path: []string{"foobar"},
				},
				{
					res: &testresource.FakeResource{
						Id: "foobar",
						Attrs: &resource.Attributes{
							"foobar": "foobar",
							"barfoo": "barfoo",
						},
					},
					path: []string{"barfoo"},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id: "foobar",
						Attrs: &resource.Attributes{
							"foobar": "foobar",
							"barfoo": "barfoo",
						},
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
					Id:   "foobar",
					Type: "fakeres",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
						"barfoo": "barfoo",
						"struct": map[string]interface{}{
							"baz": "baz",
							"bar": "bar",
						},
					},
				},
				&testresource.FakeResource{
					Id:   "barfoo",
					Type: "fakeres",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
						"barfoo": "barfoo",
						"struct": map[string]interface{}{
							"baz": "baz",
							"bar": "bar",
						},
					},
				},
				&testresource.FakeResource{
					Id:   "foobaz",
					Type: "other",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
						"barfoo": "barfoo",
						"struct": map[string]interface{}{
							"baz": "baz",
							"bar": "bar",
						},
					},
				},
				&testresource.FakeResource{
					Id:   "resource",
					Type: "other",
					Attrs: &resource.Attributes{
						"foobar": "foobar",
						"barfoo": "barfoo",
						"struct": map[string]interface{}{
							"baz": "baz",
							"bar": "bar",
						},
						"structslice": []map[string]interface{}{
							{
								"string": "one",
								"array":  []string{"foo"},
							},
						},
					},
				},
			},
			cloud: []resource.Resource{
				&testresource.FakeResource{
					Id:   "foobar",
					Type: "fakeres",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
						"barfoo": "foobar",
						"struct": map[string]interface{}{
							"baz": "bar",
							"bar": "baz",
						},
					},
				},
				&testresource.FakeResource{
					Id:   "barfoo",
					Type: "fakeres",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
						"barfoo": "foobar",
						"struct": map[string]interface{}{
							"baz": "bar",
							"bar": "baz",
						},
					},
				},
				&testresource.FakeResource{
					Id:   "foobaz",
					Type: "other",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
						"barfoo": "foobar",
						"struct": map[string]interface{}{
							"baz": "bar",
							"bar": "baz",
						},
					},
				},
				&testresource.FakeResource{
					Id:   "resource",
					Type: "other",
					Attrs: &resource.Attributes{
						"foobar": "barfoo",
						"barfoo": "foobar",
						"struct": map[string]interface{}{
							"baz": "bar",
							"bar": "baz",
						},
						"structslice": []map[string]interface{}{
							{
								"string": "two",
								"array":  []string{"oof"},
							},
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
			},
			expected: Analysis{
				managed: []resource.Resource{
					&testresource.FakeResource{
						Id:   "resource",
						Type: "other",
						Attrs: &resource.Attributes{
							"foobar": "foobar",
							"barfoo": "barfoo",
							"struct": map[string]interface{}{
								"baz": "baz",
								"bar": "bar",
							},
							"structslice": []map[string]interface{}{
								{
									"string": "one",
									"array":  []string{"foo"},
								},
							},
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
							Id:   "resource",
							Type: "other",
							Attrs: &resource.Attributes{
								"foobar": "foobar",
								"barfoo": "barfoo",
								"struct": map[string]interface{}{
									"baz": "baz",
									"bar": "bar",
								},
								"structslice": []map[string]interface{}{
									{
										"string": "one",
										"array":  []string{"foo"},
									},
								},
							},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "barfoo",
									To:   "foobar",
									Path: []string{
										"barfoo",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "foobar",
									To:   "barfoo",
									Path: []string{
										"foobar",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "bar",
									To:   "baz",
									Path: []string{
										"struct",
										"bar",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "baz",
									To:   "bar",
									Path: []string{
										"struct",
										"baz",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "foo",
									To:   "oof",
									Path: []string{
										"structslice",
										"0",
										"array",
										"0",
									},
								},
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "one",
									To:   "two",
									Path: []string{
										"structslice",
										"0",
										"string",
									},
								},
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
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestDiff with computed field send 1 alert",
			iac: []resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"id":  "ID",
						"arn": "ARN",
					},
				},
			},
			cloud: []resource.Resource{
				&resource.AbstractResource{
					Id:   "ID",
					Type: aws.AwsAmiResourceType,
					Attrs: &resource.Attributes{
						"id":  "IDCHANGED",
						"arn": "ARNCHANGED",
					},
				},
			},
			alerts: alerter.Alerts{},
			expected: Analysis{
				managed: []resource.Resource{
					&resource.AbstractResource{
						Id:   "ID",
						Type: aws.AwsAmiResourceType,
						Attrs: &resource.Attributes{
							"id":  "ID",
							"arn": "ARN",
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
						Res: &resource.AbstractResource{
							Id:   "ID",
							Type: aws.AwsAmiResourceType,
							Attrs: &resource.Attributes{
								"id":  "ID",
								"arn": "ARN",
							},
						},
						Changelog: Changelog{
							{
								Change: diff.Change{
									Type: "update",
									From: "ARN",
									To:   "ARNCHANGED",
									Path: []string{
										"arn",
									},
								},
								Computed: true,
							},
							{
								Change: diff.Change{
									Type: "update",
									From: "ID",
									To:   "IDCHANGED",
									Path: []string{
										"id",
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
				&resource.AbstractResource{
					Id:   "managed security group",
					Type: aws.AwsSecurityGroupResourceType,
					Attrs: &resource.Attributes{
						"id": "managed security group",
					},
				},
			},
			cloud: []resource.Resource{
				&resource.AbstractResource{
					Id:   "managed security group",
					Type: aws.AwsSecurityGroupResourceType,
					Attrs: &resource.Attributes{
						"id": "managed security group",
					},
				},
				&resource.AbstractResource{
					Id:   "unmanaged rule",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"id": "unmanaged rule",
					},
				},
			},
			expected: Analysis{
				managed: []resource.Resource{
					&resource.AbstractResource{
						Id:   "managed security group",
						Type: aws.AwsSecurityGroupResourceType,
						Attrs: &resource.Attributes{
							"id": "managed security group",
						},
					},
				},
				unmanaged: []resource.Resource{
					&resource.AbstractResource{
						Id:   "unmanaged rule",
						Type: aws.AwsSecurityGroupRuleResourceType,
						Attrs: &resource.Attributes{
							"id": "unmanaged rule",
						},
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

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			aws.InitResourcesMetadata(repo)

			analyzer := NewAnalyzer(al, AnalyzerOptions{Deep: true})

			for _, res := range c.cloud {
				addSchemaToRes(res, repo)
			}

			for _, res := range c.iac {
				addSchemaToRes(res, repo)
			}

			for _, res := range c.ignoredRes {
				addSchemaToRes(res, repo)
			}

			for _, drift := range c.ignoredDrift {
				addSchemaToRes(drift.res, repo)
			}

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

func addSchemaToRes(res resource.Resource, repo resource.SchemaRepositoryInterface) {
	abstractResource, ok := res.(*resource.AbstractResource)
	if ok {
		schema, _ := repo.GetSchema(res.TerraformType())
		abstractResource.Sch = schema
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
			Id:   "test-driftctl2",
			Type: "aws_iam_user",
			Attrs: &resource.Attributes{
				"foobar": "test",
			},
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
					Path: []string{"status"},
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
			&resource.SerializedResource{
				Id:   "AKIA5QYBVVD25KFXJHYJ",
				Type: "aws_iam_access_key",
			},
			&resource.SerializedResource{
				Id:   "test-managed",
				Type: "aws_iam_user",
			},
		},
		unmanaged: []resource.Resource{
			&resource.SerializedResource{
				Id:   "driftctl",
				Type: "aws_s3_bucket_policy",
			},
			&resource.SerializedResource{
				Id:   "driftctl",
				Type: "aws_s3_bucket_notification",
			},
		},
		deleted: []resource.Resource{
			&resource.SerializedResource{
				Id:   "test-driftctl2",
				Type: "aws_iam_user",
			},
			&resource.SerializedResource{
				Id:   "AKIA5QYBVVD2Y6PBAAPY",
				Type: "aws_iam_access_key",
			},
		},
		differences: []Difference{
			{
				Res: &resource.SerializedResource{
					Id:   "AKIA5QYBVVD25KFXJHYJ",
					Type: "aws_iam_access_key",
				},
				Changelog: []Change{
					{
						Change: diff.Change{
							Type: "update",
							Path: []string{"status"},
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
