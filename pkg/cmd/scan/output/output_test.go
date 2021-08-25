package output

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/alerts"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
)

func fakeAnalysis() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddUnmanaged(
		resource.NewResource(
			"unmanaged-id-1",
			"aws_unmanaged_resource",
		),
		resource.NewResource(
			"unmanaged-id-2",
			"aws_unmanaged_resource",
		),
	)
	a.AddDeleted(
		resource.NewResource(
			"deleted-id-1",
			"aws_deleted_resource",
		),
		resource.NewResource(
			"deleted-id-2",
			"aws_deleted_resource",
		),
	)
	a.AddManaged(
		resource.NewResource(
			"diff-id-1",
			"aws_diff_resource",
		),
		resource.NewResource(
			"no-diff-id-1",
			"aws_no_diff_resource",
		),
	)
	a.AddDifference(analyser.Difference{Res: resource.NewResource(
		"diff-id-1",
		"aws_diff_resource").SetSource(&resource.TerraformStateSource{
		State:  "tfstate://state.tfstate",
		Module: "module",
		Name:   "name",
	}),
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
	return &a
}

func fakeAnalysisWithAlerts() *analyser.Analysis {
	a := fakeAnalysis()
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, "aws_vpc", "aws_vpc", alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, "aws_sqs", "aws_sqs", alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, "aws_sns", "aws_sns", alerts.EnumerationPhase),
		},
	})
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisNoDrift() *analyser.Analysis {
	a := analyser.Analysis{}
	for i := 0; i < 5; i++ {
		a.AddManaged(resource.NewResource(
			"managed-id-"+fmt.Sprintf("%d", i),
			"aws_managed_resource",
		))
	}
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithJsonFields() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddManaged(
		resource.NewResource(
			"diff-id-1",
			"aws_diff_resource",
		),
	)
	a.AddManaged(
		resource.NewResource(
			"diff-id-2",
			"aws_diff_resource",
		),
	)
	a.AddDifference(analyser.Difference{
		Res: resource.NewResource(
			"diff-id-1",
			"aws_diff_resource").SetSource(&resource.TerraformStateSource{
			State:  "tfstate://state.tfstate",
			Module: "module",
			Name:   "name",
		}),
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
		Res: resource.NewResource(
			"diff-id-2",
			"aws_diff_resource").SetSource(&resource.TerraformStateSource{
			State:  "tfstate://state.tfstate",
			Module: "module",
			Name:   "name",
		}),
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
	return &a
}

func fakeAnalysisWithoutAttrs() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddDeleted(
		resource.NewResource(
			"dfjkgnbsgj",
			"FakeResourceStringer",
		).SetAttributes(&resource.Attributes{}),
	)
	a.AddManaged(
		resource.NewResource(
			"usqyfsdbgjsdgjkdfg",
			"FakeResourceStringer",
		).SetAttributes(&resource.Attributes{}),
	)
	a.AddUnmanaged(
		resource.NewResource(
			"duysgkfdjfdgfhd",
			"FakeResourceStringer",
		).SetAttributes(&resource.Attributes{}),
	)
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithStringerResources() *analyser.Analysis {
	a := analyser.Analysis{}
	schema := &resource.Schema{HumanReadableAttributesFunc: func(res *resource.Resource) map[string]string {
		return map[string]string{
			"Name": (*res.Attrs)["name"].(string),
		}
	}}
	a.AddDeleted(
		resource.NewResource(
			"dfjkgnbsgj",
			"FakeResourceStringer").SetAttributes(&resource.Attributes{
			"name": "deleted resource",
		}).SetSchema(schema),
	)
	a.AddManaged(
		resource.NewResource(
			"usqyfsdbgjsdgjkdfg",
			"FakeResourceStringer",
		).SetAttributes(&resource.Attributes{
			"name": "managed resource",
		}).SetSchema(schema),
	)
	a.AddUnmanaged(
		resource.NewResource(
			"duysgkfdjfdgfhd",
			"FakeResourceStringer",
		).SetAttributes(&resource.Attributes{
			"name": "unmanaged resource",
		}).SetSchema(schema),
	)
	a.AddDifference(analyser.Difference{
		Res: resource.NewResource(
			"gdsfhgkbn",
			"FakeResourceStringer",
		).
			SetAttributes(&resource.Attributes{
				"name": "resource with diff",
			}).
			SetSchema(schema).
			SetSource(&resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			}),
		Changelog: []analyser.Change{
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
	return &a
}

func fakeAnalysisWithComputedFields() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddManaged(
		resource.NewResource(
			"diff-id-1",
			"aws_diff_resource",
		),
	)
	a.AddDifference(analyser.Difference{
		Res: resource.NewResource(
			"diff-id-1",
			"aws_diff_resource",
		).
			SetSource(&resource.TerraformStateSource{
				State:  "tfstate://state.tfstate",
				Module: "module",
				Name:   "name",
			}),
		Changelog: []analyser.Change{
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
		},
	})
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			analyser.NewComputedDiffAlert(),
		},
	})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithAWSEnumerationError() *analyser.Analysis {
	a := analyser.Analysis{}
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, "aws_vpc", "aws_vpc", alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, "aws_sqs", "aws_sqs", alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteAWSTerraform, "aws_sns", "aws_sns", alerts.EnumerationPhase),
		},
	})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisWithGithubEnumerationError() *analyser.Analysis {
	a := analyser.Analysis{}
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, "github_team", "github_team", alerts.EnumerationPhase),
			alerts.NewRemoteAccessDeniedAlert(common.RemoteGithubTerraform, "github_team_membership", "github_team", alerts.EnumerationPhase),
		},
	})
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return &a
}

func fakeAnalysisForJSONPlan() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddUnmanaged(
		resource.NewResource(
			"unmanaged-id-1",
			"aws_unmanaged_resource").SetAttributes(&resource.Attributes{
			"name": "First unmanaged resource",
		}),
		resource.NewResource(
			"unmanaged-id-2",
			"aws_unmanaged_resource").SetAttributes(&resource.Attributes{
			"name": "Second unmanaged resource",
		}),
	)
	a.AddManaged(
		resource.NewResource(
			"managed-id-1",
			"aws_managed_resource").SetAttributes(&resource.Attributes{
			"name": "First managed resource",
		}),
		resource.NewResource(
			"managed-id-2",
			"aws_managed_resource").SetAttributes(&resource.Attributes{
			"name": "Second managed resource",
		}),
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
			want: &output.VoidPrinter{},
		},
		{
			name: "json /dev/stdout output",
			path: "/dev/stdout",
			key:  JSONOutputType,
			want: &output.VoidPrinter{},
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
			want: &output.VoidPrinter{},
		},
		{
			name: "jsonplan /dev/stdout output",
			path: "/dev/stdout",
			key:  PlanOutputType,
			want: &output.VoidPrinter{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPrinter(OutputConfig{
				Key: tt.key,
				Options: map[string]string{
					"path": tt.path,
				},
			}, tt.quiet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrinter() = %v, want %v", got, tt.want)
			}
		})
	}
}
