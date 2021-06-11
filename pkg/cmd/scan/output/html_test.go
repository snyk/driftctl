package output

import (
	"io/ioutil"
	"path"
	"testing"
	"time"

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
			name:       "test html output when there's no resources",
			goldenfile: "output_empty.html",
			analysis: func() *analyser.Analysis {
				a := &analyser.Analysis{}
				a.Date = time.Date(2021, 06, 10, 0, 0, 0, 0, &time.Location{})
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
					&testresource.FakeResource{
						Id:   "deleted-id-3",
						Type: "aws_deleted_resource",
					},
				)
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
					Id:   "diff-id-2",
					Type: "aws_diff_resource",
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
						Change: diff.Change{
							Type: diff.UPDATE,
							Path: []string{"Policies", "0"},
							From: testresource.FakeResource{},
							To:   testresource.FakeResource{Id: "093cd6ba-cf6d-4800-b252-6a50ca8903cd", Type: "aws_iam_policy"},
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

func TestHTML_PrettifyPaths(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "test simple object",
			paths:    []string{"path", "to", "field"},
			expected: "path:<br>&emsp;to:<br>&emsp;&emsp;field:",
		},
		{
			name:     "test object with array of string",
			paths:    []string{"properties", "tags", "0"},
			expected: "properties:<br>&emsp;tags:<br>&emsp;&emsp;- ",
		},
		{
			name:     "test object with array of objects",
			paths:    []string{"path", "to", "0", "field"},
			expected: "path:<br>&emsp;to:<br>&emsp;&emsp;- field:",
		},
		{
			name:     "test with simple array",
			paths:    []string{"0", "field"},
			expected: "- field:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prettifyPaths(tt.paths)
			assert.Equal(t, tt.expected, got)
		})
	}
}
