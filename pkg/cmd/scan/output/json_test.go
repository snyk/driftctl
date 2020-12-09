package output

import (
	"io/ioutil"
	"path"
	"testing"

	"github.com/cloudskiff/driftctl/test/goldenfile"

	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/pkg/analyser"
)

func TestJSON_Write(t *testing.T) {
	type args struct {
		analysis *analyser.Analysis
	}
	tests := []struct {
		name       string
		goldenfile string
		args       args
		wantErr    bool
	}{
		{
			name:       "test json output",
			goldenfile: "output.json",
			args: args{
				analysis: fakeAnalysis(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempFile, err := ioutil.TempFile(tempDir, "result")
			if err != nil {
				t.Fatal(err)
			}
			c := NewJSON(tempFile.Name())
			if err := c.Write(tt.args.analysis); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			result, err := ioutil.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}
			expectedFilePath := path.Join("./testdata/", tt.goldenfile)
			if *goldenfile.Update == tt.goldenfile {
				if err := ioutil.WriteFile(expectedFilePath, result, 0600); err != nil {
					t.Fatal(err)
				}
			}
			expected, err := ioutil.ReadFile(expectedFilePath)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(expected), string(result))
		})
	}
}
