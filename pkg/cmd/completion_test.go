package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/test"

	"github.com/spf13/cobra"
)

func TestCompletionCmd(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(NewCompletionCmd())

	tests := []struct {
		name     string
		args     []string
		expected string
		err      error
	}{
		{
			name: "Without args",
			args: []string{"completion"},
			err:  fmt.Errorf("accepts 1 arg(s), received 0"),
		},
		{
			name: "With wrong arg",
			args: []string{"completion", "test"},
			err:  fmt.Errorf("invalid argument \"test\" for \"root completion\""),
		},
		{
			name:     "With bash arg",
			args:     []string{"completion", "bash"},
			expected: "# bash completion for root",
		},
		{
			name:     "With zsh arg",
			args:     []string{"completion", "zsh"},
			expected: "#compdef _root root",
		},
		{
			name:     "With fish arg",
			args:     []string{"completion", "fish"},
			expected: "# fish completion for root",
		},
		{
			name:     "With powershell arg",
			args:     []string{"completion", "powershell"},
			expected: "Register-ArgumentCompleter -Native -CommandName 'root'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := test.Execute(rootCmd, tt.args...)

			if tt.expected != "" && !strings.Contains(output, tt.expected) {
				t.Errorf("Expected to contain: \n %v\nGot:\n %v", tt.expected, output)
			}
			if tt.err != nil && tt.err.Error() != err.Error() {
				t.Errorf("Expected %v, got %v", tt.err, err)
			}
		})
	}
}
