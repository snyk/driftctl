package cmd

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/spf13/cobra"

	"github.com/snyk/driftctl/pkg"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
)

func NewFmtCmd(opts *pkg.FmtOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "fmt",
		Long:   "Take an analysis results in JSON on stdin and return it in another format",
		Hidden: true,
		Args:   cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			outputFlag, _ := cmd.Flags().GetStringSlice("output")
			if len(outputFlag) > 1 {
				return errors.New("Only one output format can be set")
			}
			out, err := parseOutputFlags(outputFlag)
			if err != nil {
				return err
			}
			opts.Output = out[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFmt(opts, os.Stdin)
		},
	}

	fl := cmd.Flags()
	fl.StringSliceP(
		"output",
		"o",
		[]string{output.Example(output.ConsoleOutputType)},
		"Output format, by default it will write to the console\n"+
			"Accepted formats are: "+strings.Join(output.SupportedOutputsExample(), ",")+"\n",
	)

	return cmd
}

func runFmt(opts *pkg.FmtOptions, reader io.Reader) error {

	var analysisText []byte
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		analysisText = append(analysisText, scanner.Bytes()...)
	}

	analysis := analyser.NewAnalysis()
	err := json.Unmarshal(analysisText, analysis)
	if err != nil {
		return err
	}

	return output.GetOutput(opts.Output).Write(analysis)
}
