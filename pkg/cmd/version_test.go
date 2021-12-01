package cmd

import (
	"fmt"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/version"
	"github.com/cloudskiff/driftctl/test"

	"github.com/spf13/cobra"
)

func TestVersionCmd(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(NewVersionCmd())

	output, err := test.Execute(rootCmd, "version")
	if output == "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := fmt.Sprintf("%s\n", version.Current())
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}

func TestVersionCmd_Invalid(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(NewVersionCmd())

	_, err := test.Execute(rootCmd, "version", "test")
	if err == nil {
		t.Errorf("Invalid arg should generate error")
	}

	expected := `unknown command "test" for "root version"`
	if err.Error() != expected {
		t.Errorf("Expected %v, got %v", expected, err)
	}
}
