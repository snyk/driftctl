package output

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote"
	"github.com/cloudskiff/driftctl/pkg/remote/aws"
	"github.com/cloudskiff/driftctl/pkg/remote/github"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/r3labs/diff/v2"
)

func fakeAnalysis() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddUnmanaged(
		&testresource.FakeResource{
			Id:   "unmanaged-id-1",
			Type: "aws_unmanaged_resource",
		},
		&testresource.FakeResource{
			Id:   "unmanaged-id-2",
			Type: "aws_unmanaged_resource",
		},
	)
	a.AddDeleted(
		&testresource.FakeResource{
			Id:   "deleted-id-1",
			Type: "aws_deleted_resource",
		}, &testresource.FakeResource{
			Id:   "deleted-id-2",
			Type: "aws_deleted_resource",
		},
	)
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
		&testresource.FakeResource{
			Id:   "no-diff-id-1",
			Type: "aws_no_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
		Id:   "diff-id-1",
		Type: "aws_diff_resource",
	}, Changelog: []analyser.Change{
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
	return &a
}

func fakeAnalysisNoDrift() *analyser.Analysis {
	a := analyser.Analysis{}
	for i := 0; i < 5; i++ {
		a.AddManaged(&testresource.FakeResource{
			Id:   "managed-id-" + fmt.Sprintf("%d", i),
			Type: "aws_managed_resource",
		})
	}
	return &a
}

func fakeAnalysisWithJsonFields() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
	)
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-2",
			Type: "aws_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
		Id:   "diff-id-1",
		Type: "aws_diff_resource",
	}, Changelog: []analyser.Change{
		{
			Change: diff.Change{
				Type: diff.UPDATE,
				Path: []string{"Json"},
				From: "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Removed\":\"Added\",\"Changed\":[\"ec2:DescribeInstances\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}",
				To:   "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Changed\":[\"ec2:*\"],\"NewField\":[\"foobar\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}",
			},
		},
	}})
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
		Id:   "diff-id-2",
		Type: "aws_diff_resource",
	}, Changelog: []analyser.Change{
		{
			Change: diff.Change{
				Type: diff.UPDATE,
				Path: []string{"Json"},
				From: "{\"foo\":\"bar\"}",
				To:   "{\"bar\":\"foo\"}",
			},
		},
	}})
	return &a
}

func fakeAnalysisWithStringerResources() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddDeleted(
		&testresource.FakeResourceStringer{
			Id:   "dfjkgnbsgj",
			Name: "deleted resource",
		},
	)
	a.AddManaged(
		&testresource.FakeResourceStringer{
			Id:   "usqyfsdbgjsdgjkdfg",
			Name: "managed resource",
		},
	)
	a.AddUnmanaged(
		&testresource.FakeResourceStringer{
			Id:   "duysgkfdjfdgfhd",
			Name: "unmanaged resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: &testresource.FakeResourceStringer{
		Id:   "gdsfhgkbn",
		Name: "resource with diff",
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
	return &a
}

func fakeAnalysisWithComputedFields() *analyser.Analysis {
	a := analyser.Analysis{}
	a.AddManaged(
		&testresource.FakeResource{
			Id:   "diff-id-1",
			Type: "aws_diff_resource",
		},
	)
	a.AddDifference(analyser.Difference{Res: testresource.FakeResource{
		Id:   "diff-id-1",
		Type: "aws_diff_resource",
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
	return &a
}

func fakeAnalysisWithAWSEnumerationError() *analyser.Analysis {
	a := analyser.Analysis{}
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			remote.NewEnumerationAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_vpc", "aws_vpc"),
			remote.NewEnumerationAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_sqs", "aws_sqs"),
			remote.NewEnumerationAccessDeniedAlert(aws.RemoteAWSTerraform, "aws_sns", "aws_sns"),
		},
	})
	return &a
}

func fakeAnalysisWithGithubEnumerationError() *analyser.Analysis {
	a := analyser.Analysis{}
	a.SetAlerts(alerter.Alerts{
		"": []alerter.Alert{
			remote.NewEnumerationAccessDeniedAlert(github.RemoteGithubTerraform, "github_team", "github_team"),
			remote.NewEnumerationAccessDeniedAlert(github.RemoteGithubTerraform, "github_team_membership", "github_team"),
		},
	})
	return &a
}

func TestGetPrinter(t *testing.T) {
	tests := []struct {
		name string
		path string
		key  string
		want output.Printer
	}{
		{
			name: "json file output",
			path: "/path/to/file",
			key:  JSONOutputType,
			want: output.NewConsolePrinter(),
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPrinter(OutputConfig{
				Key: tt.key,
				Options: map[string]string{
					"path": tt.path,
				},
			}); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPrinter() = %v, want %v", got, tt.want)
			}
		})
	}
}
