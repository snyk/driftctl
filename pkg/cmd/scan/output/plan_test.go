package output

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/test/goldenfile"
	"github.com/stretchr/testify/assert"
)

func TestPlan_Write(t *testing.T) {
	tests := []struct {
		name       string
		goldenfile string
		analysis   *analyser.Analysis
		wantErr    bool
	}{
		{
			name:       "test jsonplan output",
			goldenfile: "output_plan.json",
			analysis:   fakeAnalysisForJSONPlan(),
			wantErr:    false,
		},
		{
			name:       "test jsonplan output when no infra",
			goldenfile: "output_plan_empty.json",
			analysis:   &analyser.Analysis{},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempFile, err := ioutil.TempFile(tempDir, "result")
			if err != nil {
				t.Fatal(err)
			}
			c := NewPlan(tempFile.Name())
			if err := c.Write(tt.analysis); (err != nil) != tt.wantErr {
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

func TestPlan_Write_stdout(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		goldenfile string
		analysis   *analyser.Analysis
		wantErr    bool
	}{
		{
			name:       "test jsonplan output on stdout",
			goldenfile: "output_plan.json",
			path:       "stdout",
			analysis:   fakeAnalysisForJSONPlan(),
			wantErr:    false,
		},

		{
			name:       "test jsonplan output on /dev/stdout",
			goldenfile: "output_plan.json",
			path:       "/dev/stdout",
			analysis:   fakeAnalysisForJSONPlan(),
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			c := NewPlan(tt.path)
			if err := c.Write(tt.analysis); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}

			outC := make(chan []byte)
			// copy the output in a separate goroutine so printing can't block indefinitely
			go func() {
				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				outC <- buf.Bytes()
			}()

			// back to normal state
			w.Close()
			os.Stdout = stdout // restoring the real stdout
			result := <-outC

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
