package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/config"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/mocks"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDriftctlCmd_Help(t *testing.T) {
	cmd := NewDriftctlCmd(mocks.MockBuild{})

	cases := []struct {
		args []string
	}{
		{args: []string{}},
		{args: []string{"help"}},
		{args: []string{"--help"}},
		{args: []string{"-h"}},
	}

	for _, tt := range cases {
		output, err := test.Execute(&cmd.Command, tt.args...)
		if output == "" {
			t.Errorf("Unexpected output: %v", output)
		}
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := cmd.UsageString()
		if !strings.Contains(output, expected) {
			t.Errorf("Expected to contain: \n %v\nGot:\n %v", expected, output)
		}
	}
}

func TestDriftctlCmd_Version(t *testing.T) {
	cmd := NewDriftctlCmd(mocks.MockBuild{})

	output, err := test.Execute(&cmd.Command, "version")
	if output == "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "dev-dev\n"
	if output != expected {
		t.Errorf("Expected %v, got %v", expected, output)
	}
}

func TestDriftctlCmd_Completion(t *testing.T) {
	cmd := NewDriftctlCmd(mocks.MockBuild{})

	output, err := test.Execute(&cmd.Command, "completion", "bash")
	if output == "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "# bash completion for driftctl"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v", expected, output)
	}
}

func TestDriftctlCmd_Scan(t *testing.T) {

	cases := []struct {
		env  map[string]string
		args []string
		err  error
	}{
		{},
		{
			env: map[string]string{
				"DCTL_TO": "test",
			},
			err: fmt.Errorf("unsupported cloud provider 'test'\nValid values are: aws+tf,github+tf"),
		},
		{
			env: map[string]string{
				"DCTL_TO": "test",
			},
			args: []string{"--to", "aws+tf"},
		},
		{
			env: map[string]string{
				"DCTL_FROM": "test",
			},
			err: fmt.Errorf("Unable to parse from flag 'test': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://"),
		},
		{
			env: map[string]string{
				"DCTL_FROM": "test",
			},
			args: []string{"--from", "tfstate://terraform.tfstate"},
		},
		{
			env: map[string]string{
				"DCTL_OUTPUT": "test",
			},
			err: fmt.Errorf("Unable to parse output flag 'test': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json"),
		},
		{
			env: map[string]string{
				"DCTL_OUTPUT": "test",
			},
			args: []string{"--output", "console://"},
		},
		{
			env: map[string]string{
				"DCTL_FILTER": "Type='test'",
			},
			err: fmt.Errorf("unable to parse filter expression: SyntaxError: Expected tRbracket, received: tUnknown"),
		},
		{
			env: map[string]string{
				"DCTL_FILTER": "Type='test'",
			},
			args: []string{"--filter", "Type=='test'"},
		},
	}

	config.Init()
	for index, c := range cases {
		t.Run(fmt.Sprintf("%d", index), func(t *testing.T) {
			if c.env != nil && len(c.env) > 0 {
				for key, val := range c.env {
					_ = os.Setenv(key, val)
					defer os.Unsetenv(key)
				}
			}
			cmd := NewDriftctlCmd(mocks.MockBuild{})
			scanCmd, _, _ := cmd.Find([]string{"scan"})
			scanCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
			args := append([]string{"scan"}, c.args...)
			_, err := test.Execute(&cmd.Command, args...)
			if c.err == nil && err != nil || c.err != nil && err == nil {
				t.Fatalf("Got error '%s', expected '%s'", err, c.err)
			}
			if c.err != nil && err != nil && err.Error() != c.err.Error() {
				t.Fatalf("Got error '%s', expected '%s'", err.Error(), c.err.Error())
			}
		})
	}
}

func TestDriftctlCmd_Invalid(t *testing.T) {
	cmd := NewDriftctlCmd(mocks.MockBuild{})

	cases := []struct {
		args     []string
		expected string
	}{
		{args: []string{"test"}, expected: `unknown command "test" for "driftctl"`},
		{args: []string{"-t"}, expected: `unknown shorthand flag: 't' in -t`},
		{args: []string{"--test"}, expected: `unknown flag: --test`},
	}

	for _, tt := range cases {
		_, err := test.Execute(&cmd.Command, tt.args...)
		if err == nil {
			t.Errorf("Invalid arg should generate error")
		}
		if err.Error() != tt.expected {
			t.Errorf("Expected %v, got %v", tt.expected, err)
		}
	}
}

func TestDriftctlCmd_ShouldCheckVersion(t *testing.T) {
	cases := []struct {
		Name      string
		IsRelease bool
		args      []string
		env       map[string]string
		expected  bool
	}{
		{
			Name:      "When we are in release mode and no args, should check for update",
			IsRelease: true,
			args:      []string{""},
			expected:  true,
		},
		{
			Name:      "Don't check for update for version cmd",
			IsRelease: true,
			args:      []string{"version"},
			expected:  false,
		},
		{
			Name:      "Don't check for update for help cmd",
			IsRelease: true,
			args:      []string{"help"},
			expected:  false,
		},
		{
			Name:      "Don't check for update for cmd --help",
			IsRelease: true,
			args:      []string{"scan", "--help"},
			expected:  false,
		},
		{
			Name:      "Don't check for update for cmd -h",
			IsRelease: true,
			args:      []string{"scan", "-h"},
			expected:  false,
		},
		{
			Name:      "Don't check for update when no check flag present",
			IsRelease: true,
			args:      []string{"--no-version-check"},
			expected:  false,
		},
		{
			Name:      "Don't check for update in dev mode",
			IsRelease: false,
			args:      []string{""},
			expected:  false,
		},
		{
			Name:      "Don't check for update when env DCTL_NO_VERSION_CHECK set",
			IsRelease: true,
			env: map[string]string{
				"DCTL_NO_VERSION_CHECK": "foo",
			},
			expected: false,
		},
		{
			Name:      "Should not return error when launching sub command",
			IsRelease: false,
			args:      []string{"scan", "--from", "tfstate://terraform.tfstate"},
			expected:  false,
		},
		{
			Name:      "Don't check for update for completion cmd",
			IsRelease: true,
			args:      []string{"completion", "bash"},
			expected:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(tt *testing.T) {
			assert := assert.New(tt)

			os.Clearenv()
			for key, val := range c.env {
				os.Setenv(key, val)
			}

			cmd := NewDriftctlCmd(mocks.MockBuild{Release: c.IsRelease})
			os.Args = append([]string{"driftctl"}, c.args...)
			result := cmd.ShouldCheckVersion()

			assert.Equal(c.expected, result)
		})
	}
}

func TestContainCmd(t *testing.T) {
	cases := []struct {
		args     []string
		cmd      string
		expected bool
	}{
		{args: []string{}, cmd: "", expected: false},
		{args: []string{"scan"}, cmd: "version", expected: false},
		{args: []string{"version"}, cmd: "version", expected: true},
	}

	for _, tt := range cases {
		if got := contains(tt.args, tt.cmd); got != tt.expected {
			t.Errorf("Expected %v, got %v", tt.expected, got)
		}
	}
}
