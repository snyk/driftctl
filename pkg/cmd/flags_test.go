package cmd

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/pkg/iac/config"
)

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
		out []string
	}
	tests := []struct {
		name string
		args args
		want []output.OutputConfig
		err  error
	}{
		{
			name: "test empty output",
			args: args{
				out: []string{""},
			},
			want: []output.OutputConfig{},
			err:  fmt.Errorf("Unable to parse output flag '': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json,plan://PATH/TO/FILE.json"),
		},
		{
			name: "test empty array",
			args: args{
				out: []string{},
			},
			want: []output.OutputConfig{},
			err:  nil,
		},
		{
			name: "test invalid",
			args: args{
				out: []string{"sdgjsdgjsdg"},
			},
			want: []output.OutputConfig{},
			err:  fmt.Errorf("Unable to parse output flag 'sdgjsdgjsdg': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json,plan://PATH/TO/FILE.json"),
		},
		{
			name: "test invalid",
			args: args{
				out: []string{"://"},
			},
			want: []output.OutputConfig{},
			err:  fmt.Errorf("Unable to parse output flag '://': \nAccepted formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json,plan://PATH/TO/FILE.json"),
		},
		{
			name: "test unsupported",
			args: args{
				out: []string{"foobar://"},
			},
			want: []output.OutputConfig{},
			err:  fmt.Errorf("Unsupported output 'foobar': \nValid formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json,plan://PATH/TO/FILE.json"),
		},
		{
			name: "test empty json",
			args: args{
				out: []string{"json://"},
			},
			want: []output.OutputConfig{},
			err:  fmt.Errorf("Invalid json output 'json://': \nMust be of kind: json://PATH/TO/FILE.json"),
		},
		{
			name: "test valid console",
			args: args{
				out: []string{"console://"},
			},
			want: []output.OutputConfig{
				{
					Key: "console",
				},
			},
			err: nil,
		},
		{
			name: "test valid json",
			args: args{
				out: []string{"json:///tmp/foobar.json"},
			},
			want: []output.OutputConfig{
				{
					Key:  "json",
					Path: "/tmp/foobar.json",
				},
			},
			err: nil,
		},
		{
			name: "test empty jsonplan",
			args: args{
				out: []string{"plan://"},
			},
			want: []output.OutputConfig{},
			err:  fmt.Errorf("Invalid plan output 'plan://': \nMust be of kind: plan://PATH/TO/FILE.json"),
		},
		{
			name: "test valid jsonplan",
			args: args{
				out: []string{"plan:///tmp/foobar.json"},
			},
			want: []output.OutputConfig{
				{
					Key:  "plan",
					Path: "/tmp/foobar.json",
				},
			},
			err: nil,
		},
		{
			name: "test multiple output values",
			args: args{
				out: []string{"console:///dev/stdout", "json://result.json"},
			},
			want: []output.OutputConfig{
				{
					Key: "console",
				},
				{
					Key:  "json",
					Path: "result.json",
				},
			},
			err: nil,
		},
		{
			name: "test multiple output values with invalid value",
			args: args{
				out: []string{"console:///dev/stdout", "invalid://result.json"},
			},
			want: []output.OutputConfig{
				{
					Key: "console",
				},
			},
			err: fmt.Errorf("Unsupported output 'invalid': \nValid formats are: console://,html://PATH/TO/FILE.html,json://PATH/TO/FILE.json,plan://PATH/TO/FILE.json"),
		},
		{
			name: "test multiple valid output values",
			args: args{
				out: []string{"json://result1.json", "json://result2.json", "json://result3.json"},
			},
			want: []output.OutputConfig{
				{
					Key:  "json",
					Path: "result1.json",
				},
				{
					Key:  "json",
					Path: "result2.json",
				},
				{
					Key:  "json",
					Path: "result3.json",
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutputFlags(tt.args.out)
			if err != nil && err.Error() != tt.err.Error() {
				t.Fatalf("got error = '%v', expected '%v'", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("parseOutputFlag() got = '%v', want '%v'", got, tt.want)
			}
		})
	}
}
