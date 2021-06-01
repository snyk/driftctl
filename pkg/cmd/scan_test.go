package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/cmd/scan/output"

	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/cloudskiff/driftctl/test"

	"github.com/spf13/cobra"
)

// TODO: Test successful scan
func TestScanCmd(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(NewScanCmd())
	// test.Execute(rootCmd, "scan")

}

func TestScanCmd_Valid(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	scanCmd := NewScanCmd()
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
		{args: []string{"scan", "-t", "glou"}, expected: "unsupported cloud provider 'glou'\nValid values are: aws+tf,github+tf"},
		{args: []string{"scan", "--to"}, expected: `flag needs an argument: --to`},
		{args: []string{"scan", "--to", "glou"}, expected: "unsupported cloud provider 'glou'\nValid values are: aws+tf,github+tf"},
		{args: []string{"scan", "-f"}, expected: `flag needs an argument: 'f' in -f`},
		{args: []string{"scan", "--from"}, expected: `flag needs an argument: --from`},
		{args: []string{"scan", "--from"}, expected: `flag needs an argument: --from`},
		{args: []string{"scan", "--from", "tosdgjhgsdhgkjs"}, expected: "Unable to parse from flag 'tosdgjhgsdhgkjs': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://"},
		{args: []string{"scan", "--from", "://"}, expected: "Unable to parse from flag '://': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://"},
		{args: []string{"scan", "--from", "://test"}, expected: "Unable to parse from flag '://test': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://"},
		{args: []string{"scan", "--from", "tosdgjhgsdhgkjs://"}, expected: "Unable to parse from flag 'tosdgjhgsdhgkjs://': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://"},
		{args: []string{"scan", "--from", "terraform+foo+bar://test"}, expected: "Unable to parse from scheme 'terraform+foo+bar': \nAccepted schemes are: tfstate://,tfstate+s3://,tfstate+http://,tfstate+https://,tfstate+tfcloud://"},
		{args: []string{"scan", "--from", "unsupported://test"}, expected: "Unsupported IaC source 'unsupported': \nAccepted values are: tfstate"},
		{args: []string{"scan", "--from", "tfstate+foobar://test"}, expected: "Unsupported IaC backend 'foobar': \nAccepted values are: s3,http,https,tfcloud"},
		{args: []string{"scan", "--from", "tfstate:///tmp/test", "--from", "tfstate+toto://test"}, expected: "Unsupported IaC backend 'toto': \nAccepted values are: s3,http,https,tfcloud"},
		{args: []string{"scan", "--filter", "Type='test'"}, expected: "unable to parse filter expression: SyntaxError: Expected tRbracket, received: tUnknown"},
		{args: []string{"scan", "--filter", "Type='test'", "--filter", "Type='test2'"}, expected: "Filter flag should be specified only once"},
		{args: []string{"scan", "--tf-provider-version", ".30.2"}, expected: "Invalid version argument .30.2, expected a valid semver string (e.g. 2.13.4)"},
		{args: []string{"scan", "--tf-provider-version", "foo"}, expected: "Invalid version argument foo, expected a valid semver string (e.g. 2.13.4)"},
	}

	for _, tt := range cases {
		rootCmd := &cobra.Command{Use: "root"}
		rootCmd.AddCommand(NewScanCmd())
		_, err := test.Execute(rootCmd, tt.args...)
		if err == nil {
			t.Errorf("Invalid arg should generate error")
		}
		if err.Error() != tt.expected {
			t.Errorf("Expected '%v', got '%v'", tt.expected, err)
		}
	}
}

func Test_parseFromFlag(t *testing.T) {
	type args struct {
		from []string
	}
	tests := []struct {
		name    string
		args    args
		want    []config.SupplierConfig
		wantErr bool
	}{
		{
			name: "test complete from parsing",
			args: args{
				from: []string{"tfstate+s3://bucket/path/to/state.tfstate"},
			},
			want: []config.SupplierConfig{
				{
					Key:     "tfstate",
					Backend: "s3",
					Path:    "bucket/path/to/state.tfstate",
				},
			},
			wantErr: false,
		},
		{
			name: "test complete from parsing with multiples flags",
			args: args{
				from: []string{"tfstate+s3://bucket/path/to/state.tfstate", "tfstate:///tmp/my-state.tfstate"},
			},
			want: []config.SupplierConfig{
				{
					Key:     "tfstate",
					Backend: "s3",
					Path:    "bucket/path/to/state.tfstate",
				},
				{
					Key:     "tfstate",
					Backend: "",
					Path:    "/tmp/my-state.tfstate",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFromFlag(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFromFlag() error = %v, err %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFromFlag() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseOutputFlag(t *testing.T) {
	type args struct {
		out string
	}
	tests := []struct {
		name string
		args args
		want *output.OutputConfig
		err  error
	}{
		{
			name: "test empty",
			args: args{
				out: "",
			},
			want: nil,
			err:  fmt.Errorf("Unable to parse output flag '': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json"),
		},
		{
			name: "test invalid",
			args: args{
				out: "sdgjsdgjsdg",
			},
			want: nil,
			err:  fmt.Errorf("Unable to parse output flag 'sdgjsdgjsdg': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json"),
		},
		{
			name: "test invalid",
			args: args{
				out: "://",
			},
			want: nil,
			err:  fmt.Errorf("Unable to parse output flag '://': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json"),
		},
		{
			name: "test unsupported",
			args: args{
				out: "foobar://",
			},
			want: nil,
			err:  fmt.Errorf("Unsupported output 'foobar': \nValid formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json"),
		},
		{
			name: "test empty json",
			args: args{
				out: "json://",
			},
			want: nil,
			err:  fmt.Errorf("Invalid json output 'json://': \nMust be of kind: json://PATH/TO/FILE.json"),
		},
		{
			name: "test valid console",
			args: args{
				out: "console://",
			},
			want: &output.OutputConfig{
				Key:     "console",
				Options: map[string]string{},
			},
			err: nil,
		},
		{
			name: "test valid json",
			args: args{
				out: "json:///tmp/foobar.json",
			},
			want: &output.OutputConfig{
				Key: "json",
				Options: map[string]string{
					"path": "/tmp/foobar.json",
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutputFlag(tt.args.out)
			if err != nil && err.Error() != tt.err.Error() {
				t.Fatalf("got error = '%v', expected '%v'", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseOutputFlag() got = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
