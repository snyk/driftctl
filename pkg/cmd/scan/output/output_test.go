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
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/output"
)

func fakeAnalysis() *analyser.Analysis {
	a := analyser.NewAnalysis()
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
	a.ProviderName = "AWS"
	a.ProviderVersion = "3.19.0"
	return a
}

func fakeAnalysisWithAlerts() *analyser.Analysis {
	a := fakeAnalysis()
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

func fakeAnalysisWithoutAttrs() *analyser.Analysis {
	a := analyser.NewAnalysis()
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
