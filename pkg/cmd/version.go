package cmd

import (
	"text/template"

	"github.com/snyk/driftctl/pkg/version"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display driftctl version",
		Long:  "Display driftctl version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			data := struct {
				Version string
			}{Version: version.Current()}
			t := template.Must(template.New("version").Parse(versionTemplate))
			err := t.Execute(cmd.OutOrStdout(), data)
			if err != nil {
				panic(err)
			}
		},
	}
	return cmd
}
