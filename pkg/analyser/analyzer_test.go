package analyser

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	dctlresource "github.com/snyk/driftctl/pkg/resource"

	alerter2 "github.com/snyk/driftctl/enumeration/alerter"

	"github.com/snyk/driftctl/pkg/filter"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	testresource "github.com/snyk/driftctl/test/resource"

	"github.com/snyk/driftctl/test/goldenfile"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"

	"github.com/r3labs/diff/v2"
)

func TestAnalyze(t *testing.T) {
	cases := []struct {
		name         string
		iac          []*resource.Resource
		ignoredRes   []*resource.Resource
		cloud        []*resource.Resource
		ignoredDrift []struct {
			res  *resource.Resource
			path []string
		}
		alerts     alerter2.Alerts
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
			iac:      []*resource.Resource{},
			cloud:    []*resource.Resource{},
			expected: Analysis{},
		},
		{
			name: "TestIgnoreFromCoverageIacNotInCloud",
			iac: []*resource.Resource{
				{
					Id: "foobar",
				},
			},
			cloud: []*resource.Resource{},
			expected: Analysis{
				summary: Summary{
					TotalResources: 1,
					TotalDeleted:   1,
				},
				deleted: []*resource.Resource{
					{
						Id: "foobar",
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "TestResourceIgnoredDeleted",
			iac: []*resource.Resource{
				{
					Id: "foobar",
				},
			},
			ignoredRes: []*resource.Resource{
				{
					Id: "foobar",
				},
			},
			cloud: []*resource.Resource{},
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
			iac: []*resource.Resource{
				{
					Id: "foobar",
				},
				{
					Id: "foobar2",
				},
			},
			ignoredRes: []*resource.Resource{
				{
					Id: "foobar2",
				},
			},
			cloud: []*resource.Resource{
				{
					Id: "foobar",
				},
				{
					Id: "foobar2",
				},
			},
			expected: Analysis{
				managed: []*resource.Resource{
					{
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
			iac: []*resource.Resource{
				{
					Id: "foobar",
				},
			},
			cloud: []*resource.Resource{
				{
					Id: "foobar",
				},
			},
			expected: Analysis{
				managed: []*resource.Resource{
					{
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
			iac:  []*resource.Resource{},
			cloud: []*resource.Resource{
				{
					Id: "foobar",
				},
			},
			expected: Analysis{
				summary: Summary{
					TotalResources: 1,
					TotalUnmanaged: 1,
				},
				unmanaged: []*resource.Resource{
					{
						Id: "foobar",
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "Test alert on unmanaged security group rules",
			iac: []*resource.Resource{
				{
					Id:   "managed security group",
					Type: aws.AwsSecurityGroupResourceType,
					Attrs: &resource.Attributes{
						"id": "managed security group",
					},
				},
			},
			cloud: []*resource.Resource{
				{
					Id:   "managed security group",
					Type: aws.AwsSecurityGroupResourceType,
					Attrs: &resource.Attributes{
						"id": "managed security group",
					},
				},
				{
					Id:   "unmanaged rule",
					Type: aws.AwsSecurityGroupRuleResourceType,
					Attrs: &resource.Attributes{
						"id": "unmanaged rule",
					},
				},
			},
			expected: Analysis{
				managed: []*resource.Resource{
					{
						Id:   "managed security group",
						Type: aws.AwsSecurityGroupResourceType,
						Attrs: &resource.Attributes{
							"id": "managed security group",
						},
					},
				},
				unmanaged: []*resource.Resource{
					{
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
				alerts: alerter2.Alerts{
					"": {
						newUnmanagedSecurityGroupRulesAlert(),
					},
				},
			},
			hasDrifted: true,
		},
		{
			name: "Test sorted unmanaged & deleted resources",
			iac: []*resource.Resource{
				{
					Id:   "deleted resource 22",
					Type: "aws_s3_bucket",
				},
				{
					Id:   "deleted resource 20",
					Type: "aws_ebs_volume",
				},
				{
					Id:   "deleted resource 20",
					Type: "aws_s3_bucket",
				},
			},
			cloud: []*resource.Resource{
				{
					Id:   "unmanaged resource 12",
					Type: "aws_s3_bucket",
				},
				{
					Id:   "unmanaged resource 10",
					Type: "aws_s3_bucket",
				},
				{
					Id:   "unmanaged resource 11",
					Type: "aws_ebs_volume",
				},
			},
			expected: Analysis{
				managed: []*resource.Resource{},
				unmanaged: []*resource.Resource{
					{
						Id:   "unmanaged resource 11",
						Type: "aws_ebs_volume",
					},
					{
						Id:   "unmanaged resource 10",
						Type: "aws_s3_bucket",
					},
					{
						Id:   "unmanaged resource 12",
						Type: "aws_s3_bucket",
					},
				},
				deleted: []*resource.Resource{
					{
						Id:   "deleted resource 20",
						Type: "aws_ebs_volume",
					},
					{
						Id:   "deleted resource 20",
						Type: "aws_s3_bucket",
					},
					{
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
				alerts: alerter2.Alerts{},
			},
			hasDrifted: true,
		},
		{
			name: "Test Discriminant function",
			iac: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsAppAutoscalingTargetResourceType,
					Attrs: &resource.Attributes{
						"scalable_dimension": "test2",
					},
				},
			},
			cloud: []*resource.Resource{
				{
					Id:   "foo",
					Type: aws.AwsAppAutoscalingTargetResourceType,
					Attrs: &resource.Attributes{
						"scalable_dimension": "test1",
					},
				},
				{
					Id:   "foo",
					Type: aws.AwsAppAutoscalingTargetResourceType,
					Attrs: &resource.Attributes{
						"scalable_dimension": "test2",
					},
				},
			},
			hasDrifted: true,
			expected: Analysis{
				managed: []*resource.Resource{
					{
						Id:   "foo",
						Type: aws.AwsAppAutoscalingTargetResourceType,
						Attrs: &resource.Attributes{
							"scalable_dimension": "test2",
						},
					},
				},
				unmanaged: []*resource.Resource{
					{
						Id:   "foo",
						Type: aws.AwsAppAutoscalingTargetResourceType,
						Attrs: &resource.Attributes{
							"scalable_dimension": "test1",
						},
					},
				},
				summary: Summary{
					TotalResources: 2,
					TotalManaged:   1,
					TotalUnmanaged: 1,
				},
			},
		},
	}

	differ, err := diff.NewDiffer(diff.SliceOrdering(true))
	if err != nil {
		t.Fatalf("Error creating new differ: %e", err)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			testFilter := &filter.MockFilter{}
			for _, ignored := range c.ignoredRes {
				testFilter.On("IsResourceIgnored", ignored).Return(true)
			}
			testFilter.On("IsResourceIgnored", mock.Anything).Return(false)

			al := alerter2.NewAlerter()
			if c.alerts != nil {
				al.SetAlerts(c.alerts)
			}

			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			aws.InitResourcesMetadata(repo)

			analyzer := NewAnalyzer(al, testFilter)

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

			result, err := analyzer.Analyze(c.cloud, c.iac)

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

func addSchemaToRes(res *resource.Resource, repo dctlresource.SchemaRepositoryInterface) {
	schema, _ := repo.GetSchema(res.ResourceType())
	res.Sch = schema
}

func TestAnalysis_MarshalJSON(t *testing.T) {
	goldenFile := "./testdata/output.json"
	analysis := Analysis{
		Duration: 241 * time.Second,
		Date:     time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC),
	}
	analysis.SetIaCSourceCount(1)
	analysis.AddManaged(
		&resource.Resource{
			Id:   "AKIA5QYBVVD25KFXJHYJ",
			Type: "aws_iam_access_key",
		}, &resource.Resource{
			Id:   "driftctl2",
			Type: "aws_managed_resource",
		},
	)
	analysis.AddUnmanaged(
		&resource.Resource{
			Id:   "driftctl",
			Type: "aws_s3_bucket_policy",
		}, &resource.Resource{
			Id:   "driftctl",
			Type: "aws_s3_bucket_notification",
		},
	)
	analysis.AddDeleted(
		&resource.Resource{
			Id:   "test-driftctl2",
			Type: "aws_iam_user",
			Attrs: &resource.Attributes{
				"foobar": "test",
			},
		},
		&resource.Resource{
			Id:   "AKIA5QYBVVD2Y6PBAAPY",
			Type: "aws_iam_access_key",
		},
	)
	analysis.SetAlerts(alerter2.Alerts{
		"aws_iam_access_key": {
			&alerter2.FakeAlert{Msg: "This is an alert"},
		},
	})
	analysis.ProviderName = "AWS"
	analysis.ProviderVersion = "2.18.5"

	got, err := json.MarshalIndent(analysis, "", "\t")
	if err != nil {
		t.Fatal(err)
	}
	if *goldenfile.Update == "TestAnalysis_MarshalJSON" {
		if err := os.WriteFile(goldenFile, got, 0600); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := os.ReadFile(goldenFile)
	if err != nil {
		t.Fatal(err)
	}
	assert.Nil(t, err)
	assert.Equal(t, string(expected), string(got))
}

func TestAnalysis_UnmarshalJSON(t *testing.T) {
	expected := Analysis{
		summary: Summary{
			TotalResources:      6,
			TotalUnmanaged:      2,
			TotalDeleted:        2,
			TotalManaged:        2,
			TotalIaCSourceCount: 3,
		},
		managed: []*resource.Resource{
			{
				Id:   "AKIA5QYBVVD25KFXJHYJ",
				Type: "aws_iam_access_key",
			},
			{
				Id:   "test-managed",
				Type: "aws_iam_user",
			},
		},
		unmanaged: []*resource.Resource{
			{
				Id:   "driftctl",
				Type: "aws_s3_bucket_policy",
			},
			{
				Id:   "driftctl",
				Type: "aws_s3_bucket_notification",
			},
		},
		deleted: []*resource.Resource{
			{
				Id:   "test-driftctl2",
				Type: "aws_iam_user",
			},
			{
				Id:   "AKIA5QYBVVD2Y6PBAAPY",
				Type: "aws_iam_access_key",
			},
		},
		alerts: alerter2.Alerts{
			"aws_iam_access_key": {
				&alerter2.SerializedAlert{
					Msg: "This is an alert",
				},
			},
		},
		ProviderName:    "AWS",
		ProviderVersion: "2.18.5",
		Date:            time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC),
	}

	got := Analysis{}
	input, err := os.ReadFile("./testdata/input.json")
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
	assert.Equal(t, uint(3), got.Summary().TotalIaCSourceCount)
	assert.Len(t, got.alerts, 1)
	assert.Equal(t, got.alerts["aws_iam_access_key"][0].Message(), "This is an alert")
}
