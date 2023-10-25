package output

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/test/goldenfile"
)

func TestHTML_Write(t *testing.T) {
	tests := []struct {
		name       string
		goldenfile string
		analysis   func() *analyser.Analysis
		err        error
	}{
		{
			name:       "test html output when there's no resources",
			goldenfile: "output_empty.html",
			analysis: func() *analyser.Analysis {
				a := &analyser.Analysis{}
				a.Date = time.Date(2021, 06, 10, 0, 0, 0, 0, &time.Location{})
				a.ProviderName = "AWS"
				a.ProviderVersion = "3.19.0"
				return a
			},
			err: nil,
		},
		{
			name:       "test html output when infrastructure is in sync",
			goldenfile: "output_sync.html",
			analysis: func() *analyser.Analysis {
				a := &analyser.Analysis{}
				a.Date = time.Date(2021, 06, 10, 0, 0, 0, 0, &time.Location{})
				a.Duration = 72 * time.Second
				a.AddManaged(
					&resource.Resource{
						Id:   "deleted-id-3",
						Type: "aws_deleted_resource",
					},
				)
				a.ProviderName = "AWS"
				a.ProviderVersion = "3.19.0"
				return a
			},
			err: nil,
		},
		{
			name:       "test html output",
			goldenfile: "output.html",
			analysis: func() *analyser.Analysis {
				a := fakeAnalysisWithAlerts()
				a.Date = time.Date(2021, 06, 10, 0, 0, 0, 0, &time.Location{})
				a.Duration = 91 * time.Second
				a.AddManaged(
					&resource.Resource{
						Id:   "diff-id-2",
						Type: "aws_diff_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://state.tfstate",
							Name:   "diff-id-2",
							Module: "module",
						},
					},
					&resource.Resource{
						Id:   "diff-id-3",
						Type: "aws_diff_resource",
						Source: &resource.TerraformStateSource{
							State: "tfstate+s3://state2.tfstate",
							Name:  "b",
						},
					},
				)
				a.AddDeleted(
					&resource.Resource{
						Id:   "deleted-id-3",
						Type: "aws_deleted_resource",
						Source: &resource.TerraformStateSource{
							State: "tfstate://deleted/terraform.tfstate",
							Name:  "deleted-id-3",
						},
					},
					&resource.Resource{
						Id:   "deleted-id-4",
						Type: "aws_deleted_resource",
						Source: &resource.TerraformStateSource{
							State: "tfstate://deleted/terraform.tfstate",
							Name:  "deleted-id-3",
						},
					},
					&resource.Resource{
						Id:   "deleted-id-5",
						Type: "aws_deleted_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://deleted/terraform.tfstate",
							Name:   "deleted-id-3",
							Module: "module-1",
						},
					},
					&resource.Resource{
						Id:   "deleted-id-6",
						Type: "aws_deleted_resource",
					},
				)
				a.AddUnmanaged(
					&resource.Resource{
						Id:   "unmanaged-id-3",
						Type: "aws_unmanaged_resource",
					},
					&resource.Resource{
						Id:   "unmanaged-id-4",
						Type: "aws_unmanaged_resource",
					},
					&resource.Resource{
						Id:   "unmanaged-id-5",
						Type: "aws_unmanaged_resource",
					},
				)
				a.ProviderName = "AWS"
				a.ProviderVersion = "3.19.0"
				return a
			},
			err: nil,
		},
		{
			name:       "test html output when coverage is 100",
			goldenfile: "output_coverage_100.html",
			analysis: func() *analyser.Analysis {
				a := &analyser.Analysis{}
				a.Date = time.Date(2021, 06, 10, 0, 0, 0, 0, &time.Location{})
				a.Duration = 91 * time.Second
				a.AddManaged(
					&resource.Resource{
						Id:   "resource-id-1",
						Type: "aws_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://state.tfstate",
							Module: "module",
							Name:   "name",
						},
					},
				)
				a.ProviderName = "AWS"
				a.ProviderVersion = "3.19.0"
				return a
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "result")

			if err != nil {
				t.Fatal(err)
			}
			c := NewHTML(tempFile.Name())

			err = c.Write(tt.analysis())
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}

			got, err := os.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}

			expectedFilePath := path.Join("./testdata/", tt.goldenfile)
			if *goldenfile.Update == tt.goldenfile {
				if err := os.WriteFile(expectedFilePath, got, 0600); err != nil {
					t.Fatal(err)
				}
			}

			expected, err := os.ReadFile(expectedFilePath)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, string(expected), string(got))
		})
	}
}

func TestHTML_DistinctResourceTypes(t *testing.T) {
	tests := []struct {
		name      string
		resources []*resource.Resource
		value     []string
	}{
		{
			name:      "should return empty array",
			resources: []*resource.Resource{},
			value:     []string{},
		},
		{
			name: "should return distinct list of resource types",
			resources: []*resource.Resource{
				{
					Id:   "deleted-id-1",
					Type: "aws_deleted_resource",
				},
				{
					Id:   "unmanaged-id-1",
					Type: "aws_unmanaged_resource",
				},
				{
					Id:   "unmanaged-id-2",
					Type: "aws_unmanaged_resource",
				},
				{
					Id:   "diff-id-1",
					Type: "aws_diff_resource",
				},
				{
					Id:   "deleted-id-2",
					Type: "aws_deleted_resource",
				},
			},
			value: []string{"aws_deleted_resource", "aws_unmanaged_resource", "aws_diff_resource"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := distinctResourceTypes(tt.resources)
			assert.Equal(t, tt.value, got)
		})
	}
}

func TestHTML_DistinctIaCSources(t *testing.T) {
	tests := []struct {
		name      string
		resources []*resource.Resource
		value     []string
	}{
		{
			name:      "should return empty array",
			resources: []*resource.Resource{},
			value:     []string{},
		},
		{
			name: "should return distinct list of iac sources",
			resources: []*resource.Resource{
				{
					Id:   "deleted-id-1",
					Type: "aws_deleted_resource",
					Source: &resource.TerraformStateSource{
						Module: "module",
						Name:   "test",
						State:  "tfstate://terraform.tfstate",
					},
				},
				{
					Id:   "unmanaged-id-1",
					Type: "aws_unmanaged_resource",
					Source: &resource.TerraformStateSource{
						Module: "module",
						Name:   "test",
						State:  "tfstate://terraform2.tfstate",
					},
				},
				{
					Id:   "unmanaged-id-2",
					Type: "aws_unmanaged_resource",
					Source: &resource.TerraformStateSource{
						Module: "module",
						Name:   "test",
						State:  "tfstate+s3://test/terraform.tfstate",
					},
				},
				{
					Id:   "diff-id-1",
					Type: "aws_diff_resource",
					Source: &resource.TerraformStateSource{
						Module: "module",
						Name:   "test",
						State:  "tfstate://terraform.tfstate",
					},
				},
				{
					Id:   "deleted-id-2",
					Type: "aws_deleted_resource",
				},
			},
			value: []string{"tfstate://terraform.tfstate", "tfstate://terraform2.tfstate", "tfstate+s3://test/terraform.tfstate"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := distinctIaCSources(tt.resources)
			assert.Equal(t, tt.value, got)
		})
	}
}
