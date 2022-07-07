package output

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote/alerts"
	"github.com/snyk/driftctl/enumeration/remote/common"
	remoteerr "github.com/snyk/driftctl/enumeration/remote/error"

	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/output"
)

func fakeAnalysis(opts analyser.AnalyzerOptions) *analyser.Analysis {
	if opts == (analyser.AnalyzerOptions{}) {
		opts = analyser.AnalyzerOptions{Deep: true}
	}
	a := analyser.NewAnalysis(opts)
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.SetIaCSourceCount(3)
	a.Duration = 12 * time.Second
	a.AddUnmanaged(
		&resource.Resource{
			Id:   "unmanaged-id-1",
			Type: "aws_unmanaged_resource",
		},
		&resource.Resource{
			Id:   "unmanaged-id-2",
			Type: "aws_unmanaged_resource",
		},
	)
	a.AddDeleted(
		&resource.Resource{
			Id:   "deleted-id-1",
			Type: "aws_deleted_resource",
			Source: &resource.TerraformStateSource{
				State:  "tfstate://delete_state.tfstate",
				Module: "module",
				Name:   "name",
			},
		}, &resource.Resource{
			Id:   "deleted-id-2",
			Type: "aws_deleted_resource",
		},
	)
	a.AddManaged(
		&resource.Resource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
		&resource.Resource{
			Id:   "no-diff-id-1",
			Type: "aws_no_diff_resource",
		},
	)
	// Cover the case when a diff occur on a resource without a source
	a.AddDifference(analyser.Difference{
		Res: &resource.Resource{
			Id:   "diff-id-2",
			Type: "aws_diff_resource",
		},
		Changelog: []analyser.Change{
			{
				Change: diff.Change{
					Type: diff.UPDATE,
					Path: []string{"updated", "field"},
					From: "foobar",
					To:   "barfoo",
				},
			},
		},
	})
	a.AddDifference(analyser.Difference{Res: &resource.Resource{
		Id:   "diff-id-1",
		Type: "aws_diff_resource",
		Source: &resource.TerraformStateSource{
			State:  "tfstate://state.tfstate",
			Module: "module",
			Name:   "name",
		},
	},
		Changelog: []analyser.Change{
			{
				Change: diff.Change{
					Type: diff.UPDATE,
					Path: []string{"updated", "field"},
					From: "foobar",
					To:   "barfoo",
				},
			},
			{
				Change: diff.Change{
					Type: diff.CREATE,
					Path: []string{"new", "field"},
					From: nil,
					To:   "newValue",
				},
			},
			{
				Change: diff.Change{
					Type: diff.DELETE,
					Path: []string{"a"},
					From: "oldValue",
					To:   nil,
				},
			},
		}})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisWithAlerts() *analyser.Analysis {
	a := fakeAnalysis(analyser.AnalyzerOptions{})
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "aws_vpc", "aws_vpc"), alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "aws_sqs", "aws_sqs"), alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "aws_sns", "aws_sns"), alerts.EnumerationPhase),
		},
	})
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisNoDrift() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		a.AddManaged(&resource.Resource{
			Id:   "managed-id-" + fmt.Sprintf("%d", i),
			Type: "aws_managed_resource",
		})
	}
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithJsonFields() *analyser.Analysis {
	a := analyser.NewAnalysis(analyser.AnalyzerOptions{Deep: true})
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.AddManaged(
		&resource.Resource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
	)
	a.AddManaged(
		&resource.Resource{
			Id:   "diff-id-2",
			Type: "aws_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{
		Res: &resource.Resource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
			Source: &resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			},
		},
		Changelog: []analyser.Change{
			{
				JsonString: true,
				Change: diff.Change{
					Type: diff.UPDATE,
					Path: []string{"Json"},
					From: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Removed\":\"Added\",\"Changed\":[\"oldValue1\", \"oldValue2\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}",
					To:   "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Changed\":\"newValue\",\"NewField\":[\"foobar\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}",
				},
			},
		}})
	a.AddDifference(analyser.Difference{
		Res: &resource.Resource{
			Id:   "diff-id-2",
			Type: "aws_diff_resource",
			Source: &resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			},
		},
		Changelog: []analyser.Change{
			{
				JsonString: true,
				Change: diff.Change{
					Type: diff.UPDATE,
					Path: []string{"Json"},
					From: "{\"foo\":\"bar\"}",
					To:   "{\"bar\":\"foo\"}",
				},
			},
		}})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisWithoutAttrs() *analyser.Analysis {
	a := analyser.NewAnalysis(analyser.AnalyzerOptions{Deep: true})
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.AddDeleted(
		&resource.Resource{
			Id:    "dfjkgnbsgj",
			Type:  "FakeResourceStringer",
			Attrs: &resource.Attributes{},
			Source: &resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			},
		},
	)
	a.AddManaged(
		&resource.Resource{
			Id:    "usqyfsdbgjsdgjkdfg",
			Type:  "FakeResourceStringer",
			Attrs: &resource.Attributes{},
		},
	)
	a.AddUnmanaged(
		&resource.Resource{
			Id:    "duysgkfdjfdgfhd",
			Type:  "FakeResourceStringer",
			Attrs: &resource.Attributes{},
		},
	)
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisWithStringerResources() *analyser.Analysis {
	a := analyser.NewAnalysis(analyser.AnalyzerOptions{Deep: true})
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	schema := &resource.Schema{HumanReadableAttributesFunc: func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": (*res.Attrs)["name"].(string),
		}
	}}
	a.AddDeleted(
		&resource.Resource{
			Id:   "dfjkgnbsgj",
			Type: "FakeResourceStringer",
			Sch:  schema,
			Attrs: &resource.Attributes{
				"name": "deleted resource",
			},
			Source: &resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			},
		},
	)
	a.AddManaged(
		&resource.Resource{
			Id:   "usqyfsdbgjsdgjkdfg",
			Type: "FakeResourceStringer",
			Sch:  schema,
			Attrs: &resource.Attributes{
				"name": "managed resource",
			},
		},
	)
	a.AddUnmanaged(
		&resource.Resource{
			Id:   "duysgkfdjfdgfhd",
			Type: "FakeResourceStringer",
			Sch:  schema,
			Attrs: &resource.Attributes{
				"name": "unmanaged resource",
			},
		},
	)
	a.AddDifference(analyser.Difference{Res: &resource.Resource{
		Id:   "gdsfhgkbn",
		Type: "FakeResourceStringer",
		Sch:  schema,
		Attrs: &resource.Attributes{
			"name": "resource with diff",
		},
		Source: &resource.TerraformStateSource{
			State:  "tfstate://state.tfstate",
			Module: "module",
			Name:   "name",
		},
	}, Changelog: []analyser.Change{
		{
			Change: diff.Change{
				Type: diff.UPDATE,
				Path: []string{"Name"},
				From: "",
				To:   "resource with diff",
			},
		},
	}})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisWithComputedFields() *analyser.Analysis {
	a := analyser.NewAnalysis(analyser.AnalyzerOptions{Deep: true})
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.AddManaged(
		&resource.Resource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{
		Res: &resource.Resource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
			Source: &resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			},
		}, Changelog: []analyser.Change{
			{
				Change: diff.Change{
					Type: diff.UPDATE,
					Path: []string{"updated", "field"},
					From: "foobar",
					To:   "barfoo",
				},
				Computed: true,
			},
			{
				Change: diff.Change{
					Type: diff.CREATE,
					Path: []string{"new", "field"},
					From: nil,
					To:   "newValue",
				},
			},
			{
				Change: diff.Change{
					Type: diff.DELETE,
					Path: []string{"a"},
					From: "oldValue",
					To:   nil,
				},
				Computed: true,
			},
			{
				Change: diff.Change{
					Type: diff.UPDATE,
					From: "foo",
					To:   "oof",
					Path: []string{
						"struct",
						"0",
						"array",
						"0",
					},
				},
				Computed: true,
			},
			{
				Change: diff.Change{
					Type: diff.UPDATE,
					From: "one",
					To:   "two",
					Path: []string{
						"struct",
						"0",
						"string",
					},
				},
				Computed: true,
			},
		}})
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			analyser.NewComputedDiffAlert(),
		},
	})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisWithAWSEnumerationError() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "aws_vpc", "aws_vpc"), alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "aws_sqs", "aws_sqs"), alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "aws_sns", "aws_sns"), alerts.EnumerationPhase),
		},
	})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithGithubEnumerationError() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "github_team", "github_team"), alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, remoteerr.NewResourceListingErrorWithType(errors.New("dummy error"), "github_team_membership", "github_team"), alerts.EnumerationPhase),
		},
	})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisForJSONPlan() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.AddUnmanaged(
		&resource.Resource{
			Id:   "unmanaged-id-1",
			Type: "aws_unmanaged_resource",
			Attrs: &resource.Attributes{
				"name": "First unmanaged resource",
			},
		},
		&resource.Resource{
			Id:   "unmanaged-id-2",
			Type: "aws_unmanaged_resource",
			Attrs: &resource.Attributes{
				"name": "Second unmanaged resource",
			},
		},
	)
	a.AddManaged(
		&resource.Resource{
			Id:   "managed-id-1",
			Type: "aws_managed_resource",
			Attrs: &resource.Attributes{
				"name": "First managed resource",
			},
		},
		&resource.Resource{
			Id:   "managed-id-2",
			Type: "aws_managed_resource",
			Attrs: &resource.Attributes{
				"name": "Second managed resource",
			},
		},
	)
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithoutDeep() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.AddUnmanaged(
		&resource.Resource{
			Id:   "unmanaged-id-1",
			Type: "aws_unmanaged_resource",
			Attrs: &resource.Attributes{
				"name": "First unmanaged resource",
			},
		},
	)
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithOnlyManagedFlag() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.SetOptions(analyser.AnalyzerOptions{
		OnlyManaged: true,
		Deep:        true,
	})
	a.AddManaged(
		&resource.Resource{
			Id:   "foo",
			Type: aws.AwsInstanceResourceType,
			Attrs: &resource.Attributes{
				"instance_type": "test2",
			},
		},
	)
	a.AddDifference(
		analyser.Difference{
			Res: &resource.Resource{
				Id:   "foo",
				Type: aws.AwsInstanceResourceType,
				Attrs: &resource.Attributes{
					"instance_type": "test2",
				},
			},
			Changelog: []analyser.Change{
				{
					Change: diff.Change{
						Type: "update",
						From: "test2",
						To:   "test1",
						Path: []string{
							"instance_type",
						},
					},
				},
			},
		})
	a.AddDeleted(
		&resource.Resource{
			Id:   "baz",
			Type: aws.AwsInstanceResourceType,
			Attrs: &resource.Attributes{
				"instance_type": "test3",
			},
		},
	)
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithOnlyUnmanagedFlag() *analyser.Analysis {
	a := analyser.Analysis{}
	a.Date = time.Date(2022, 4, 8, 10, 35, 0, 0, time.UTC)
	a.SetOptions(analyser.AnalyzerOptions{
		OnlyUnmanaged: true,
	})
	a.AddManaged(
		&resource.Resource{
			Id:   "foo",
			Type: aws.AwsInstanceResourceType,
			Attrs: &resource.Attributes{
				"instance_type": "test2",
			},
		},
	)
	a.AddUnmanaged(
		&resource.Resource{
			Id:   "bar",
			Type: aws.AwsInstanceResourceType,
			Attrs: &resource.Attributes{
				"instance_type": "test2",
			},
		},
	)
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func TestGetPrinter(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		key   string
		quiet bool
		want  output.Printer
	}{
		{
			name: "json file output",
			path: "/path/to/file",
			key:  JSONOutputType,
			want: output.NewConsolePrinter(),
		},
		{
			name:  "json file output quiet",
			path:  "/path/to/file",
			key:   JSONOutputType,
			quiet: true,
			want:  &output.VoidPrinter{},
		},
		{
			name: "json stdout output",
			path: "stdout",
			key:  JSONOutputType,
			want: &output.ConsolePrinter{},
		},
		{
			name: "json /dev/stdout output",
			path: "/dev/stdout",
			key:  JSONOutputType,
			want: &output.ConsolePrinter{},
		},
		{
			name: "console stdout output",
			path: "stdout",
			key:  ConsoleOutputType,
			want: output.NewConsolePrinter(),
		},
		{
			name:  "quiet console stdout output",
			path:  "stdout",
			quiet: true,
			key:   ConsoleOutputType,
			want:  &output.VoidPrinter{},
		},
		{
			name: "jsonplan file output",
			path: "/path/to/file",
			key:  PlanOutputType,
			want: output.NewConsolePrinter(),
		},
		{
			name: "jsonplan stdout output",
			path: "stdout",
			key:  PlanOutputType,
			want: &output.ConsolePrinter{},
		},
		{
			name: "jsonplan /dev/stdout output",
			path: "/dev/stdout",
			key:  PlanOutputType,
			want: &output.ConsolePrinter{},
		},
		{
			name: "html stdout output",
			path: "stdout",
			key:  HTMLOutputType,
			want: &output.ConsolePrinter{},
		},
		{
			name: "html /dev/stdout output",
			path: "/dev/stdout",
			key:  HTMLOutputType,
			want: &output.ConsolePrinter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPrinter(OutputConfig{
				Key:  tt.key,
				Path: tt.path,
			}, tt.quiet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrinter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldPrint(t *testing.T) {
	tests := []struct {
		name    string
		outputs []OutputConfig
		quiet   bool
		want    bool
	}{
		{
			name: "test stdout should not prevents printing",
			outputs: []OutputConfig{
				{
					Path: "stdout",
					Key:  JSONOutputType,
				},
			},
			want: true,
		},
		{
			name: "test output to file doesn't prevent printing",
			outputs: []OutputConfig{
				{
					Path: "result.json",
					Key:  JSONOutputType,
				},
			},
			want: true,
		},
		{
			name: "test quiet should prevents printing",
			outputs: []OutputConfig{
				{
					Path: "result.json",
					Key:  JSONOutputType,
				},
			},
			quiet: true,
			want:  false,
		},
		{
			name: "test stdout should not prevents printing",
			outputs: []OutputConfig{
				{
					Path: "result.json",
					Key:  JSONOutputType,
				},
				{
					Path: "stdout",
					Key:  PlanOutputType,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldPrint(tt.outputs, tt.quiet); got != tt.want {
				t.Errorf("ShouldPrint() = %v, want %v", got, tt.want)
			}
		})
	}
}
