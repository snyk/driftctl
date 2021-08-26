package output

import (
	"io/ioutil"
	"path"
	"testing"
	"time"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/test/goldenfile"
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
				a.AddDeleted(
					&resource.Resource{
						Id:   "deleted-id-3",
						Type: "aws_deleted_resource",
					},
					&resource.Resource{
						Id:   "deleted-id-4",
						Type: "aws_deleted_resource",
					},
					&resource.Resource{
						Id:   "deleted-id-5",
						Type: "aws_deleted_resource",
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
				a.AddDifference(analyser.Difference{
					Res: &resource.Resource{
						Id:   "diff-id-2",
						Type: "aws_diff_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://state.tfstate",
							Module: "module",
							Name:   "name",
						},
					}, Changelog: []analyser.Change{
						{
							Change: diff.Change{
								Type: diff.DELETE,
								Path: []string{"path", "to", "fields", "0"},
								From: "value",
								To:   nil,
							},
						},
						{
							Change: diff.Change{
								Type: diff.UPDATE,
								Path: []string{"path", "to", "fields", "1"},
								From: 12,
								To:   "12",
							},
						},
						{
							Change: diff.Change{
								Type: diff.DELETE,
								Path: []string{"group_ids"},
								From: []string{"a071314398026"},
								To:   nil,
							},
						},
						{
							JsonString: true,
							Change: diff.Change{
								Type: diff.UPDATE,
								Path: []string{"policy"},
								From: `{"Statement":[{"Action":["s3:GetObjectVersion"],"Effect":"Allow","Principal":"*","Resource":"arn:aws:s3:::tmxxrn.foobar.driftctl-test.com/*","Sid":"PublicRead"}],"Version":"2012-10-17"}`,
								To:   `{"Statement":[{"Action":["*"],"Effect":"Deny","Principal":"*","Resource":"arn:aws:s3:::tmxxrn.foobar.driftctl-test.com/b/*","Sid":"PublicReadWrite"},{"Effect":"Deny","Sid":"PublicReadWrite"}],"Version":"2021-10-17","Test":[]}`,
							},
						},
						{
							Change: diff.Change{
								Type: diff.CREATE,
								Path: []string{"Tags", "0", "Name"},
								From: nil,
								To:   "test",
							},
						},
						{
							Change: diff.Change{
								Type: diff.UPDATE,
								Path: []string{"InstanceInitiatedShutdownBehavior"},
								From: "",
								To:   nil,
							},
						},
					}})
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
					},
				)
				a.AddDifference(analyser.Difference{
					Res: &resource.Resource{
						Id:   "resource-id-1",
						Type: "aws_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://state.tfstate",
							Module: "module",
							Name:   "name",
						},
					}, Changelog: []analyser.Change{
						{
							Change: diff.Change{
								Type: diff.DELETE,
								Path: []string{"path", "to", "fields", "0"},
								From: "value",
								To:   nil,
							},
						},
					}})
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
			tempFile, err := ioutil.TempFile(tempDir, "result")

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

			got, err := ioutil.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}

			expectedFilePath := path.Join("./testdata/", tt.goldenfile)
			if *goldenfile.Update == tt.goldenfile {
				if err := ioutil.WriteFile(expectedFilePath, got, 0600); err != nil {
					t.Fatal(err)
				}
			}

			expected, err := ioutil.ReadFile(expectedFilePath)
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
			name:      "test empty array",
			resources: []*resource.Resource{},
			value:     []string{},
		},
		{
			name: "test empty array",
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
