package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewGenDriftIgnoreCmd() *cobra.Command {
	opts := &analyser.GenDriftIgnoreOptions{}

	cmd := &cobra.Command{
		Use:   "gen-driftignore",
		Short: "Generate a .driftignore file based on your scan result",
		Long:  "This command will generate a new .driftignore file containing your current drifts and send output to /dev/stdout\n\nExample: driftctl scan -o json://stdout | driftctl gen-driftignore -i /dev/stdin > .driftignore",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.InputPath == "" {
				return errors.New("Error: you must specify an input to parse JSON from. Use driftctl gen-driftignore -i <drifts.json>\nGenerate a JSON file using the output flag: driftctl scan -o json://path/to/drifts.json")
			}

			_, list, err := genDriftIgnore(opts)
			if err != nil {
				return err
			}

			fmt.Println(list)

			return nil
		},
	}

	fl := cmd.Flags()

	fl.BoolVar(&opts.ExcludeUnmanaged, "exclude-unmanaged", false, "Exclude resources not managed by IaC")
	fl.BoolVar(&opts.ExcludeDeleted, "exclude-missing", false, "Exclude missing resources")
	fl.BoolVar(&opts.ExcludeDrifted, "exclude-changed", false, "Exclude resources that changed on cloud provider")
	fl.StringVarP(&opts.InputPath, "input", "i", "", "Input where the JSON should be parsed from")

	return cmd
}

func genDriftIgnore(opts *analyser.GenDriftIgnoreOptions) (int, string, error) {
	input, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return 0, "", err
	}

	analysis := &analyser.Analysis{}
	err = json.Unmarshal(input, analysis)
	if err != nil {
		return 0, "", err
	}

	n, list := analysis.DriftIgnoreList(*opts)

	return n, list, nil
}
