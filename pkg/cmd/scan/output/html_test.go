package output

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	testresource "github.com/cloudskiff/driftctl/test/resource"
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
			name:       "test html output",
			goldenfile: "output.html",

			analysis: func() *analyser.Analysis {
				a := fakeAnalysisWithAlerts()
				a.AddDeleted(
					&testresource.FakeResource{
						Id:   "deleted-id-3",
						Type: "aws_deleted_resource",
					},
					&testresource.FakeResource{
						Id:   "deleted-id-4",
						Type: "aws_deleted_resource",
					},
					&testresource.FakeResource{
						Id:   "deleted-id-5",
						Type: "aws_deleted_resource",
					},
					&testresource.FakeResource{
						Id:   "deleted-id-6",
						Type: "aws_deleted_resource",
					},
				)
				a.AddUnmanaged(
					&testresource.FakeResource{
						Id:   "unmanaged-id-3",
						Type: "aws_unmanaged_resource",
					},
					&testresource.FakeResource{
						Id:   "unmanaged-id-4",
						Type: "aws_unmanaged_resource",
					},
					&testresource.FakeResource{
						Id:   "unmanaged-id-5",
						Type: "aws_unmanaged_resource",
					},
				)
				a.AddDifference(analyser.Difference{Res: &testresource.FakeResource{
					Id:   "diff-id-1",
					Type: "aws_diff_resource",
				}, Changelog: []analyser.Change{
					{
						Change: diff.Change{
							Type: diff.DELETE,
							Path: []string{"path", "to", "field"},
							From: nil,
							To:   []string{"value"},
						},
					},
				}})

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
		resources []resource.Resource
		value     []string
	}{
		{
			name:      "test empty array",
			resources: []resource.Resource{},
			value:     []string{},
		},
		{
			name: "test empty array",
			resources: []resource.Resource{
				&testresource.FakeResource{
					Id:   "deleted-id-1",
					Type: "aws_deleted_resource",
				},
				&testresource.FakeResource{
					Id:   "unmanaged-id-1",
					Type: "aws_unmanaged_resource",
				},
				&testresource.FakeResource{
					Id:   "unmanaged-id-2",
					Type: "aws_unmanaged_resource",
				},
				&testresource.FakeResource{
					Id:   "diff-id-1",
					Type: "aws_diff_resource",
				},
				&testresource.FakeResource{
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
