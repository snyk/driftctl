package cmd

import (
	"testing"

	"github.com/snyk/driftctl/pkg"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
	"github.com/snyk/driftctl/test"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TODO: Test successful scan
func TestScanCmd(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(NewScanCmd(&pkg.ScanOptions{}))
	// test.Execute(rootCmd, "scan")

}

func TestScanCmd_Valid(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	scanCmd := NewScanCmd(&pkg.ScanOptions{})
	scanCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
	rootCmd.AddCommand(scanCmd)

	cases := []struct {
		args []string
	}{
		{args: []string{"scan"}},
		{args: []string{"scan", "-t", "aws+tf"}},
		{args: []string{"scan", "--to", "aws+tf"}},
		{args: []string{"scan", "-f", "tfstate://test"}},
		{args: []string{"scan", "--from", "tfstate://test"}},
		{args: []string{"scan", "--from", "tfstate://test", "--from", "tfstate://test2"}},
		{args: []string{"scan", "-t", "aws+tf", "-f", "tfstate://test"}},
		{args: []string{"scan", "--to", "aws+tf", "--from", "tfstate://test"}},
		{args: []string{"scan", "--to", "aws+tf", "--from", "tfstate+https://github.com/state.tfstate"}},
		{args: []string{"scan", "--to", "aws+tf", "--from", "tfstate+tfcloud://workspace_id"}},
		{args: []string{"scan", "--tfc-token", "token"}},
		{args: []string{"scan", "--filter", "Type=='aws_s3_bucket'"}},
		{args: []string{"scan", "--strict"}},
		{args: []string{"scan", "--tf-provider-version", "1.2.3"}},
		{args: []string{"scan", "--tf-provider-version", "3.30.2"}},
		{args: []string{"scan", "--driftignore", "./path/to/driftignore.s3"}},
		{args: []string{"scan", "--driftignore", ".driftignore"}},
		{args: []string{"scan", "-o", "html://result.html", "-o", "json://result.json"}},
		{args: []string{"scan", "--tf-lockfile", "../.terraform.lock.hcl"}},
		{args: []string{"scan", "--only-unmanaged"}},
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

func TestScanCmd_Invalid(t *testing.T) {
	cases := []struct {
		args     []string
		expected string
	}{
		{args: []string{"scan", "test"}, expected: `unknown command "test" for "root scan"`},
		{args: []string{"scan", "-e"}, expected: `unknown shorthand flag: 'e' in -e`},
		{args: []string{"scan", "--error"}, expected: `unknown flag: --error`},
		{args: []string{"scan", "-t"}, expected: `flag needs an argument: 't' in -t`},
		{args: []string{"scan", "-t", "glou"}, expected: "unsupported cloud provider 'glou'\nValid values are: aws+tf,github+tf,gcp+tf,azure+tf"},
		{args: []string{"scan", "--to"}, expected: `flag needs an argument: --to`},
		{args: []string{"scan", "--to", "glou"}, expected: "unsupported cloud provider 'glou'\nValid values are: aws+tf,github+tf,gcp+tf,azure+tf"},
		{args: []string{"scan", "-f"}, expected: `flag needs an argument: 'f' in -f`},
		{args: []string{"scan", "--from"}, expected: `flag needs an argument: --from`},
		{args: []string{"scan", "--from"}, expected: `flag needs an argument: --from`},
		{args: []string{"scan", "--from", "tosdgjhgsdhgkjs"}, expected: "Unable to parse from flag 'tosdgjhgsdhgkjs': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://,tfstate+gs://,tfstate+azurerm://"},
		{args: []string{"scan", "--from", "://"}, expected: "Unable to parse from flag '://': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://,tfstate+gs://,tfstate+azurerm://"},
		{args: []string{"scan", "--from", "://test"}, expected: "Unable to parse from flag '://test': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://,tfstate+gs://,tfstate+azurerm://"},
		{args: []string{"scan", "--from", "tosdgjhgsdhgkjs://"}, expected: "Unable to parse from flag 'tosdgjhgsdhgkjs://': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://,tfstate+gs://,tfstate+azurerm://"},
		{args: []string{"scan", "--from", "terraform+foo+bar://test"}, expected: "Unable to parse from scheme 'terraform+foo+bar': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://,tfstate+gs://,tfstate+azurerm://"},
		{args: []string{"scan", "--from", "unsupported://test"}, expected: "Unsupported IaC source 'unsupported': \nAccepted values are: tfstate"},
		{args: []string{"scan", "--from", "tfstate+foobar://test"}, expected: "Unsupported IaC backend 'foobar': \nAccepted values are: s3,http,https,tfcloud,gs,azurerm"},
		{args: []string{"scan", "--from", "tfstate:///tmp/test", "--from", "tfstate+toto://test"}, expected: "Unsupported IaC backend 'toto': \nAccepted values are: s3,http,https,tfcloud,gs,azurerm"},
		{args: []string{"scan", "--filter", "Type='test'"}, expected: "unable to parse filter expression: SyntaxError: Expected tRbracket, received: tUnknown"},
		{args: []string{"scan", "--filter", "Type='test'", "--filter", "Type='test2'"}, expected: "Filter flag should be specified only once"},
		{args: []string{"scan", "--tf-provider-version", ".30.2"}, expected: "Invalid version argument .30.2, expected a valid semver string (e.g. 2.13.4)"},
		{args: []string{"scan", "--tf-provider-version", "foo"}, expected: "Invalid version argument foo, expected a valid semver string (e.g. 2.13.4)"},
		{args: []string{"scan", "--driftignore"}, expected: "flag needs an argument: --driftignore"},
		{args: []string{"scan", "--tf-lockfile"}, expected: "flag needs an argument: --tf-lockfile"},
	}

	for _, tt := range cases {
		rootCmd := &cobra.Command{Use: "root"}
		rootCmd.AddCommand(NewScanCmd(&pkg.ScanOptions{}))
		_, err := test.Execute(rootCmd, tt.args...)
		if err == nil {
			t.Errorf("Invalid arg should generate error")
		}
		if err.Error() != tt.expected {
			t.Errorf("Expected '%v', got '%v'", tt.expected, err)
		}
	}
}

func Test_Options(t *testing.T) {
	cases := []struct {
		name          string
		args          []string
		assertOptions func(*testing.T, *pkg.ScanOptions)
	}{
		{
			name: "lockfile should be ignored by tf-provider-version flag",
			args: []string{"scan", "--to", "aws+tf", "--tf-lockfile", "testdata/terraform_valid.lock.hcl", "--tf-provider-version", "3.41.0"},
			assertOptions: func(t *testing.T, opts *pkg.ScanOptions) {
				assert.Equal(t, "3.41.0", opts.ProviderVersion)
			},
		},
		{
			name: "should get provider version from lockfile",
			args: []string{"scan", "--to", "aws+tf", "--tf-lockfile", "testdata/terraform_valid.lock.hcl"},
			assertOptions: func(t *testing.T, opts *pkg.ScanOptions) {
				assert.Equal(t, "3.47.0", opts.ProviderVersion)
			},
		},
		{
			name: "should not find provider version in lockfile",
			args: []string{"scan", "--to", "gcp+tf", "--tf-lockfile", "testdata/terraform_valid.lock.hcl"},
			assertOptions: func(t *testing.T, opts *pkg.ScanOptions) {
				assert.Equal(t, "", opts.ProviderVersion)
			},
		},
		{
			name: "should fail to read lockfile with silent error",
			args: []string{"scan", "--to", "gcp+tf", "--tf-lockfile", "testdata/terraform_invalid.lock.hcl"},
			assertOptions: func(t *testing.T, opts *pkg.ScanOptions) {
				assert.Equal(t, "", opts.ProviderVersion)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			opts := &pkg.ScanOptions{}

			rootCmd := &cobra.Command{Use: "root"}
			scanCmd := NewScanCmd(opts)
			scanCmd.RunE = func(_ *cobra.Command, args []string) error { return nil }
			rootCmd.AddCommand(scanCmd)

			_, err := test.Execute(rootCmd, tt.args...)
			assert.NoError(t, err)
			tt.assertOptions(t, opts)
		})
	}
}

func Test_RetrieveBackendsFromHCL(t *testing.T) {
	cases := []struct {
		name     string
		dir      string
		expected []config.SupplierConfig
		wantErr  error
	}{
		{
			name: "should parse s3 backend and ignore invalid file",
			dir:  "testdata/backend/s3",
			expected: []config.SupplierConfig{
				{
					Key:     state.TerraformStateReaderSupplier,
					Backend: backend.BackendKeyS3,
					Path:    "terraform-state-prod/network/terraform.tfstate",
				},
			},
		},
		{
			name:     "should not find any match and return empty slice",
			dir:      "testdata/backend",
			expected: []config.SupplierConfig{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			configs, err := retrieveBackendsFromHCL(tt.dir)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.expected, configs)
		})
	}
}
