package cmd

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/build"
	"github.com/snyk/driftctl/pkg"
	"github.com/snyk/driftctl/sentry"
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
				err := bindEnvToFlags(cmd)
				if err != nil {
					return err
				}
				return handleReporting(cmd)
			},
			Long:          "Detect, track and alert on infrastructure drift.",
			SilenceErrors: true,
			SilenceUsage:  true,
		},
		build,
	}

	cmd.SetVersionTemplate(versionTemplate)
	cmd.AddCommand(NewVersionCmd())

	cmd.AddCommand(NewCompletionCmd())

	cmd.SetUsageTemplate(usageTemplate)

	cmd.PersistentFlags().BoolP("help", "h", false, "Display help for command")
	if cmd.build.IsUsageReportingEnabled() {
		cmd.PersistentFlags().BoolP("no-version-check", "", false, "Disable the version check")
		cmd.PersistentFlags().BoolP("disable-telemetry", "", false, "Telemetry has been removed, this flag does nothing but is left here to avoid breaking workflows.")
	}
	cmd.PersistentFlags().BoolP("send-crash-report", "", false, "Enable error reporting. Crash data will be sent to us via Sentry.\nWARNING: may leak sensitive data (please read the documentation for more details)\nThis flag should be used only if an error occurs during execution")

	cmd.AddCommand(NewScanCmd(&pkg.ScanOptions{}))
	cmd.AddCommand(NewFmtCmd(&pkg.FmtOptions{}))
	cmd.AddCommand(NewGenDriftIgnoreCmd())

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
	_, noVersionCheckEnv := os.LookupEnv("DCTL_NO_VERSION_CHECK")
	noVersionCheckVal := contains(os.Args[1:], "--no-version-check")
	hasVersionCmd := contains(os.Args[1:], "version")
	hasCompletionCmd := contains(os.Args[1:], "completion")
	isHelp := contains(os.Args[1:], "help") || contains(os.Args[1:], "--help") || contains(os.Args[1:], "-h")
	return driftctlCmd.build.IsRelease() && driftctlCmd.build.IsUsageReportingEnabled() && !hasVersionCmd && !hasCompletionCmd && !noVersionCheckVal && !isHelp && !noVersionCheckEnv
}

func IsReportingEnabled(cmd *cobra.Command) bool {
	enableReporting, err := cmd.Flags().GetBool("send-crash-report")
	if err != nil {
		return false
	}
	return enableReporting
}

func handleReporting(cmd *cobra.Command) error {
	if IsReportingEnabled(cmd) {
		return sentry.Initialize()
	}
	return nil
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
		// Ignore some global flags
		// no-version-check is ignored because we don't use cmd flags to retrieve flag in version check function
		// as we check version before cmd, we use os.Args
		if f.Name == "help" || f.Name == "no-version-check" {
			return
		}
		envKey := strings.ReplaceAll(f.Name, "-", "_")
		// Apply the viper config value to the flag when the flag is not set and viper has a value
		// Allow flags precedence over env variables
		if !f.Changed && viper.IsSet(envKey) {
			envVal := viper.GetString(envKey)
			err = cmd.Flags().Set(f.Name, envVal)
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
