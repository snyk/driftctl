package cmd

import (
	"os"
	"strings"

	"github.com/cloudskiff/driftctl/build"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var usageTemplate = `Usage: {{.UseLine}}{{if .HasAvailableSubCommands}}

COMMANDS:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name 24 }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

FLAGS:
{{ .LocalFlags.FlagUsages | trimTrailingWhitespaces }}{{end}}{{if .HasAvailableInheritedFlags}}

INHERITED FLAGS:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableSubCommands}}

LEARN MORE:
  Use "{{.CommandPath}} <command> --help" for more information about a command{{end}}
`

var versionTemplate = `{{ printf "%s\n" .Version }}`

type DriftctlCmd struct {
	cobra.Command
	build build.BuildInterface
}

func NewDriftctlCmd(build build.BuildInterface) *DriftctlCmd {
	cmd := &DriftctlCmd{
		cobra.Command{
			Use:   "driftctl <command> [flags]",
			Short: "Driftctl CLI",
			PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
				return bindEnvToFlags(cmd)
			},
			Long:          "Detect, track and alert on infrastructure drift.",
			SilenceErrors: true,
		},
		build,
	}

	cmd.SetVersionTemplate(versionTemplate)
	cmd.AddCommand(NewVersionCmd())

	cmd.SetUsageTemplate(usageTemplate)

	cmd.PersistentFlags().BoolP("help", "h", false, "Display help for command")
	cmd.PersistentFlags().BoolP("no-version-check", "", false, "Disable the version check")

	cmd.AddCommand(NewScanCmd())

	return cmd
}

func contains(args []string, cmd string) bool {
	for _, arg := range args {
		if arg == cmd {
			return true
		}
	}
	return false
}

func (driftctlCmd DriftctlCmd) ShouldCheckVersion() bool {
	noVersionCheckVal := contains(os.Args[1:], "--no-version-check")
	hasVersionCmd := contains(os.Args[1:], "version")
	isHelp := contains(os.Args[1:], "help") || contains(os.Args[1:], "--help") || contains(os.Args[1:], "-h")
	return driftctlCmd.build.IsRelease() && !hasVersionCmd && !noVersionCheckVal && !isHelp
}

// Iterate over command flags
// If the command flag is not manually set (f.Changed) we override its value
// from the according env value
func bindEnvToFlags(cmd *cobra.Command) error {
	var err error
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err != nil {
			return
		}
		envKey := strings.ReplaceAll(f.Name, "-", "_")
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		// Allow flags precedence over env variables
		if !f.Changed && viper.IsSet(envKey) {
			envVal := viper.GetString(envKey)
			err = cmd.Flags().Set(envKey, envVal)
			if err != nil {
				return
			}
			logrus.WithFields(logrus.Fields{
				"env":   envKey,
				"flag":  f.Name,
				"value": envVal,
			}).Debug("Bound environment variable to flag")
		}
	})

	return err
}
