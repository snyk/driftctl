package output

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/test/goldenfile"
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
		{
			name:       "test json output with AWS enumeration alerts",
			goldenfile: "output_access_denied_alert_aws.json",
			args: args{
				analysis: fakeAnalysisWithAWSEnumerationError(),
			},
			wantErr: false,
		},
		{
			name:       "test json output with Github enumeration alerts",
			goldenfile: "output_access_denied_alert_github.json",
			args: args{
				analysis: fakeAnalysisWithGithubEnumerationError(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempFile, err := os.CreateTemp(tempDir, "result")
			if err != nil {
				t.Fatal(err)
			}
			c := NewJSON(tempFile.Name())
			if err := c.Write(tt.args.analysis); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
			result, err := os.ReadFile(tempFile.Name())
			if err != nil {
				t.Fatal(err)
			}
			expectedFilePath := path.Join("./testdata/", tt.goldenfile)
			if *goldenfile.Update == tt.goldenfile {
				if err := os.WriteFile(expectedFilePath, result, 0600); err != nil {
					t.Fatal(err)
				}
			}
			expected, err := os.ReadFile(expectedFilePath)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(expected), string(result))
		})
	}
}

func TestJSON_Write_stdout(t *testing.T) {
	type args struct {
		analysis *analyser.Analysis
	}
	tests := []struct {
		name       string
		path       string
		goldenfile string
		args       args
		wantErr    bool
	}{
		{
			name:       "test json output stdout",
			goldenfile: "output.json",
			path:       "stdout",
			args: args{
				analysis: fakeAnalysis(),
			},
			wantErr: false,
		},

		{
			name:       "test json output /dev/stdout",
			goldenfile: "output.json",
			path:       "/dev/stdout",
			args: args{
				analysis: fakeAnalysis(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			stdout := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			c := NewJSON(tt.path)
			if err := c.Write(tt.args.analysis); (err != nil) != tt.wantErr {
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
				if err := os.WriteFile(expectedFilePath, result, 0600); err != nil {
					t.Fatal(err)
				}
			}
			expected, err := os.ReadFile(expectedFilePath)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(expected), string(result))
		})
	}
}

func TestJSON_WriteMultiplesTimesInSameFile(t *testing.T) {
	emptyAnalysis := &analyser.Analysis{}
	longerAnalysis := fakeAnalysis()
	tempDir := t.TempDir()
	tempFile, err := os.CreateTemp(tempDir, "result")
	if err != nil {
		t.Fatal(err)
	}
	c := NewJSON(tempFile.Name())

	if err := c.Write(longerAnalysis); err != nil {
		t.Errorf("First write error = %v", err)
	}

	if err := c.Write(emptyAnalysis); err != nil {
		t.Errorf("Second write error = %v", err)
	}

	result, err := os.ReadFile(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	goldenFileName := "output_multiples_times.json"
	expectedFilePath := path.Join("./testdata/", goldenFileName)
	if *goldenfile.Update == goldenFileName {
		if err := os.WriteFile(expectedFilePath, result, 0600); err != nil {
			t.Fatal(err)
		}
	}
	expected, err := os.ReadFile(expectedFilePath)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, string(expected), string(result))
}
