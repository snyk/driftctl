package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/snyk/driftctl/pkg"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_runFmt_InvalidInput(t *testing.T) {
	opts := &pkg.FmtOptions{
		Output: output.OutputConfig{
			Key: output.ConsoleOutputType,
		},
	}

	input, err := os.Open("testdata/fmt/input_stdin_invalid.json")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()

	err = runFmt(opts, input)
	require.NotNil(t, err)
	assert.Equal(t, "invalid character 'i' looking for beginning of value", err.Error())
}

func Test_runFmt(t *testing.T) {
	opts := &pkg.FmtOptions{
		Output: output.OutputConfig{
			Key: output.ConsoleOutputType,
		},
	}

	input, err := os.Open("testdata/fmt/input_stdin_valid.json")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()

	stdout := os.Stdout // keep backup of the real stdout
	stderr := os.Stderr // keep backup of the real stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	err = runFmt(opts, input)

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
	output := <-outC

	if err != nil {
		t.Fatal(err)
	}

	expectedBytes, err := os.ReadFile("testdata/fmt/expected_console.txt")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, string(expectedBytes), string(output))
}

func TestFmtCmd_Valid(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	scanCmd := NewFmtCmd(&pkg.FmtOptions{})
	scanCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
	rootCmd.AddCommand(scanCmd)

	cases := []struct {
		args []string
	}{
		{args: []string{"fmt"}},
		{args: []string{"fmt", "-o", "json://test.json"}},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			output, err := test.Execute(rootCmd, tt.args...)
			if output != "" {
				t.Errorf("Unexpected output: %v", output)
			}
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFmtCmd_Invalid(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
	}{
		{args: []string{"fmt", "test"}, expected: `unknown command "test" for "root fmt"`},
		{args: []string{"fmt", "-o", "json://test.json", "-o", "html://test.html"}, expected: "Only one output format can be set"},
		{args: []string{"fmt", "-o", "foobar://barfoo"}, expected: "Unsupported output 'foobar': \nValid formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json,plan://PATH/TO/FILE.json"},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(NewFmtCmd(&pkg.FmtOptions{}))
			_, err := test.Execute(rootCmd, tt.args...)
			if err == nil {
				t.Errorf("Invalid arg should generate error")
			}
			if err.Error() != tt.expected {
				t.Errorf("Expected '%v', got '%v'", tt.expected, err)
			}
		})
	}
}
