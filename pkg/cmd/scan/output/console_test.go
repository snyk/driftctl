package output

import (
	"bytes"
	"io"
	"os"
	"path"
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/test/goldenfile"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"

	"github.com/snyk/driftctl/pkg/analyser"
)

func TestConsole_Write(t *testing.T) {
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
			name:       "test console output",
			goldenfile: "output.txt",
			args: args{analysis: func() *analyser.Analysis {
				a := fakeAnalysis()
				a.AddDeleted(
					&resource.Resource{
						Id:   "test-id-1",
						Type: "aws_test_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://test_state.tfstate",
							Module: "module",
							Name:   "name",
						},
					},
					&resource.Resource{
						Id:   "test-id-2",
						Type: "aws_test_resource",
						Source: &resource.TerraformStateSource{
							State:  "tfstate://test_state.tfstate",
							Module: "module",
							Name:   "name",
						},
					},
				)
				a.AddUnmanaged(
					&resource.Resource{
						Id:   "test-id-1",
						Type: "aws_testing_resource",
					},
					&resource.Resource{
						Id:   "test-id-2",
						Type: "aws_resource",
					},
				)
				return a
			}()},
			wantErr: false,
		},
		{
			name:       "test console output no drift",
			goldenfile: "output_no_drift.txt",
			args:       args{analysis: fakeAnalysisNoDrift()},
			wantErr:    false,
		},
		{
			name:       "test console output with resource without attributes",
			goldenfile: "output_empty_attributes.txt",
			args:       args{analysis: fakeAnalysisWithoutAttrs()},
			wantErr:    false,
		},
		{
			name:       "test console output with AWS enumeration alerts",
			goldenfile: "output_access_denied_alert_aws.txt",
			args:       args{analysis: fakeAnalysisWithAWSEnumerationError()},
			wantErr:    false,
		},
		{
			name:       "test console output with Github enumeration alerts",
			goldenfile: "output_access_denied_alert_github.txt",
			args:       args{analysis: fakeAnalysisWithGithubEnumerationError()},
			wantErr:    false,
		},
		{
			name:       "test console output without deep mode",
			goldenfile: "output_without_deep.txt",
			args:       args{analysis: fakeAnalysisWithoutDeep()},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := testresource.InitFakeSchemaRepository("aws", "3.19.0")
			aws.InitResourcesMetadata(repo)

			c := NewConsole()

			stdout := os.Stdout // keep backup of the real stdout
			stderr := os.Stderr // keep backup of the real stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

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
			assert.Nil(t, w.Close())
			os.Stdout = stdout // restoring the real stdout
			os.Stderr = stderr
			out := <-outC

			expectedFilePath := path.Join("./testdata", tt.goldenfile)
			if *goldenfile.Update == tt.goldenfile {
				if err := os.WriteFile(expectedFilePath, out, 0600); err != nil {
					t.Fatal(err)
				}
			}

			expected, err := os.ReadFile(expectedFilePath)
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, string(expected), string(out))
		})
	}
}
