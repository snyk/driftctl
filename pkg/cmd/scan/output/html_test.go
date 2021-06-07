package output

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	testresource "github.com/cloudskiff/driftctl/test/resource"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/test/goldenfile"
)

func TestHTML_Write(t *testing.T) {
	type args struct {
		analysis *analyser.Analysis
	}
	tests := []struct {
		name       string
		goldenfile string
		args       args
		err        error
	}{
		{
			name:       "test html output",
			goldenfile: "output.html",
			args: args{
				analysis: fakeAnalysisWithAlerts(),
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

			err = c.Write(tt.args.analysis)
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
				testresource.FakeResource{
					Id:   "deleted-id-1",
					Type: "aws_deleted_resource",
				},
				testresource.FakeResource{
					Id:   "unmanaged-id-1",
					Type: "aws_unmanaged_resource",
				},
				testresource.FakeResource{
					Id:   "unmanaged-id-2",
					Type: "aws_unmanaged_resource",
				},
				testresource.FakeResource{
					Id:   "diff-id-1",
					Type: "aws_diff_resource",
				},
				testresource.FakeResource{
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
