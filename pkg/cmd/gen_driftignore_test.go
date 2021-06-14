package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGenDriftIgnoreCmd_Input(t *testing.T) {
	cases := []struct {
		name            string
		args            []string
		output          string
		existingContent string
		err             error
	}{
		{
			name:   "test error on invalid input",
			args:   []string{"-i", "./testdata/input_stdin_invalid.json"},
			output: "./testdata/output_stdin_empty.txt",
			err:    errors.New("invalid character 'i' looking for beginning of value"),
		},
		{
			name:   "test empty driftignore with valid input",
			args:   []string{"-i", "./testdata/input_stdin_empty.json"},
			output: "./testdata/output_stdin_empty.txt",
			err:    nil,
		},
		{
			name:   "test driftignore content with valid input",
			args:   []string{"-i", "./testdata/input_stdin_valid.json"},
			output: "./testdata/output_stdin_valid.txt",
			err:    nil,
		},
		{
			name:   "test driftignore content with valid input and filter missing & changed only",
			args:   []string{"-i", "./testdata/input_stdin_valid.json", "--exclude-unmanaged"},
			output: "./testdata/output_stdin_valid_filter.txt",
			err:    nil,
		},
		{
			name:   "test driftignore content with valid input and filter unmanaged only",
			args:   []string{"-i", "./testdata/input_stdin_valid.json", "--exclude-missing", "--exclude-changed"},
			output: "./testdata/output_stdin_valid_filter2.txt",
			err:    nil,
		},
		{
			name:   "test error when input file does not exist",
			args:   []string{"-i", "doesnotexist"},
			output: "./testdata/output_stdin_valid_filter2.txt",
			err:    errors.New("open doesnotexist: no such file or directory"),
		},
		{
			name:   "test error when input flag is not specified",
			args:   []string{},
			output: "",
			err:    errors.New("Error: you must specify an input to parse JSON from. Use driftctl gen-driftignore -i <drifts.json>\nGenerate a JSON file using the output flag: driftctl scan -o json://path/to/drifts.json"),
		},
		{
			name:            "test driftignore content append to existing content",
			args:            []string{"-i", "./testdata/input_stdin_valid.json"},
			output:          "./testdata/output_stdin_valid_append.txt",
			existingContent: "aws_sqs_queue.https://sqs\\.us-east-1\\.amazonaws\\.com/141177182257/queue-6qaidu\\.fifo",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rootCmd := &cobra.Command{Use: "root"}
			rootCmd.AddCommand(NewGenDriftIgnoreCmd())

			stdout := os.Stdout // keep backup of the real stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			if c.existingContent != "" {
				_, err := w.WriteString(c.existingContent)
				assert.Nil(t, err)
			}

			args := append([]string{"gen-driftignore"}, c.args...)

			_, err := test.Execute(rootCmd, args...)
			if c.err != nil {
				assert.EqualError(t, err, c.err.Error())
				return
			} else {
				assert.Equal(t, c.err, err)
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
			result := <-outC

			if c.output != "" {
				output, err := os.ReadFile(c.output)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, string(output), string(result))
			}
		})
	}
}

func TestGenDriftIgnoreCmd_ValidFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	genDriftIgnoreCmd := NewGenDriftIgnoreCmd()
	genDriftIgnoreCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
	rootCmd.AddCommand(genDriftIgnoreCmd)

	cases := []struct {
		args []string
	}{
		{args: []string{"gen-driftignore", "--exclude-unmanaged"}},
		{args: []string{"gen-driftignore", "--exclude-missing"}},
		{args: []string{"gen-driftignore", "--exclude-changed"}},
		{args: []string{"gen-driftignore", "--exclude-changed=false", "--exclude-missing=false", "--exclude-unmanaged=true"}},
		{args: []string{"gen-driftignore", "--input", "/dev/stdin"}},
		{args: []string{"gen-driftignore", "-i", "/dev/stdout"}},
	}

	for _, tt := range cases {
		output, err := test.Execute(rootCmd, tt.args...)
		if output != "" {
			t.Errorf("Unexpected output: %v", output)
		}
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	}
}

func TestGenDriftIgnoreCmd_InvalidFlags(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	genDriftIgnoreCmd := NewGenDriftIgnoreCmd()
	genDriftIgnoreCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
	rootCmd.AddCommand(genDriftIgnoreCmd)

	cases := []struct {
		args []string
		err  error
	}{
		{args: []string{"gen-driftignore", "--deleted"}, err: errors.New("unknown flag: --deleted")},
		{args: []string{"gen-driftignore", "--drifted"}, err: errors.New("unknown flag: --drifted")},
		{args: []string{"gen-driftignore", "--changed"}, err: errors.New("unknown flag: --changed")},
		{args: []string{"gen-driftignore", "--missing"}, err: errors.New("unknown flag: --missing")},
		{args: []string{"gen-driftignore", "--input"}, err: errors.New("flag needs an argument: --input")},
		{args: []string{"gen-driftignore", "-i"}, err: errors.New("flag needs an argument: 'i' in -i")},
	}

	for _, tt := range cases {
		_, err := test.Execute(rootCmd, tt.args...)
		assert.EqualError(t, err, tt.err.Error())
	}
}
