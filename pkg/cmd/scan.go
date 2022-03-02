package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/build"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/memstore"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/telemetry"
	"github.com/snyk/driftctl/pkg/terraform/lock"
	"github.com/spf13/cobra"

	"github.com/snyk/driftctl/pkg"
	"github.com/snyk/driftctl/pkg/alerter"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/supplier"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
	globaloutput "github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/terraform"
)

func NewScanCmd(opts *pkg.ScanOptions) *cobra.Command {
	opts.BackendOptions = &backend.Options{}

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan",
		Long:  "Scan",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetStringSlice("from")

			iacSource, err := parseFromFlag(from)
			if err != nil {
				return err
			}

			opts.From = iacSource

			to, _ := cmd.Flags().GetString("to")
			if !remote.IsSupported(to) {
				return errors.Errorf(
					"unsupported cloud provider '%s'\nValid values are: %s",
					to,
					strings.Join(remote.GetSupportedRemotes(), ","),
				)
			}

			GCPScope, _ := cmd.Flags().GetStringSlice("gcp-scope")
			limitScope := make([]string, 0)

			if to == common.RemoteGoogleTerraform && len(GCPScope) > 0 {
				limitScope, err = parseScopeFlag(GCPScope)
				if err != nil {
					return err
				}
			} else if to != common.RemoteGoogleTerraform && len(GCPScope) > 0 {
				return errors.New("gcp-scope can only be utilized when using " + common.RemoteGoogleTerraform + " flag")
			} else if to == common.RemoteGoogleTerraform && len(GCPScope) == 0 {
				return errors.New("gcp-scope must be specified when using " + common.RemoteGoogleTerraform + " flag")
			}

			opts.GCPScope = limitScope

			outputFlag, _ := cmd.Flags().GetStringSlice("output")

			out, err := parseOutputFlags(outputFlag)
			if err != nil {
				return err
			}
			opts.Output = out

			filterFlag, _ := cmd.Flags().GetStringArray("filter")

			if len(filterFlag) > 1 {
				return errors.New("Filter flag should be specified only once")
			}

			if len(filterFlag) == 1 && filterFlag[0] != "" {
				expr, err := filter.BuildExpression(filterFlag[0])
				if err != nil {
					return errors.Wrap(err, "unable to parse filter expression")
				}
				opts.Filter = expr
			}

			providerVersion, _ := cmd.Flags().GetString("tf-provider-version")
			if err := validateTfProviderVersionString(providerVersion); err != nil {
				return err
			}
			opts.ProviderVersion = providerVersion

			if opts.ProviderVersion == "" {
				lockfilePath, _ := cmd.Flags().GetString("tf-lockfile")

				// Attempt to read the provider version from a terraform lock file
				lockFile, err := lock.ReadLocksFromFile(lockfilePath)
				if err != nil {
					logrus.WithField("error", err.Error()).Debug("Error while parsing terraform lock file")
				}
				if provider := lockFile.GetProviderByAddress(common.RemoteParameter(to).GetProviderAddress()); provider != nil {
					opts.ProviderVersion = provider.Version
					logrus.WithFields(logrus.Fields{"version": opts.ProviderVersion, "provider": to}).Debug("Found provider version in terraform lock file")
				}
			}

			opts.Quiet, _ = cmd.Flags().GetBool("quiet")
			opts.DisableTelemetry, _ = cmd.Flags().GetBool("disable-telemetry")

			opts.ConfigDir, _ = cmd.Flags().GetString("config-dir")

			if onlyManaged, _ := cmd.Flags().GetBool("only-managed"); onlyManaged {
				opts.Deep = true
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return scanRun(opts)
		},
	}

	warn := color.New(color.FgYellow, color.Bold).SprintfFunc()

	fl := cmd.Flags()
	fl.Bool(
		"quiet",
		false,
		"Do not display anything but scan results",
	)
	fl.StringSlice(
		"gcp-scope",
		[]string{},
		"Set the GCP scope for search",
	)
	fl.StringArray(
		"filter",
		[]string{},
		"JMESPath expression to filter on\n"+
			"Examples : \n"+
			"  - Type == 'aws_s3_bucket' (will filter only s3 buckets)\n"+
			"  - Type =='aws_s3_bucket && Id != 'my_bucket' (excludes s3 bucket 'my_bucket')\n"+
			"  - Attr.Tags.Terraform == 'true' (include only resources that have Tag Terraform equal to 'true')\n",
	)
	fl.StringSliceP(
		"output",
		"o",
		[]string{output.Example(output.ConsoleOutputType)},
		"Output format, by default it will write to the console\n"+
			"Accepted formats are: "+strings.Join(output.SupportedOutputsExample(), ",")+"\n",
	)
	fl.StringSliceP(
		"from",
		"f",
		[]string{"tfstate://terraform.tfstate"},
		"IaC sources, by default try to find local terraform.tfstate file\n"+
			"Accepted schemes are: "+strings.Join(supplier.GetSupportedSchemes(), ",")+"\n",
	)
	supportedRemotes := remote.GetSupportedRemotes()
	fl.StringVarP(
		&opts.To,
		"to",
		"t",
		supportedRemotes[0],
		"Cloud provider source\n"+
			"Accepted values are: "+strings.Join(supportedRemotes, ",")+"\n",
	)
	fl.StringToStringVarP(&opts.BackendOptions.Headers,
		"headers",
		"H",
		map[string]string{},
		"Use those HTTP headers to query the provided URL.\n"+
			"Only used with tfstate+http(s) backend for now.\n",
	)
	fl.StringVar(&opts.BackendOptions.TFCloudToken,
		"tfc-token",
		"",
		"Terraform Cloud / Enterprise API token.\n"+
			"Only used with tfstate+tfcloud backend.\n",
	)
	fl.StringVar(&opts.BackendOptions.TFCloudEndpoint,
		"tfc-endpoint",
		"https://app.terraform.io/api/v2",
		"Terraform Cloud / Enterprise API endpoint.\n"+
			"Only used with tfstate+tfcloud backend.\n",
	)
	fl.StringVar(&opts.BackendOptions.AzureRMBackendOptions.StorageAccount,
		"azurerm-storage-account",
		os.Getenv("AZURE_STORAGE_ACCOUNT"),
		"Azure storage account name for state backend.\n",
	)
	fl.StringVar(&opts.BackendOptions.AzureRMBackendOptions.StorageKey,
		"azurerm-account-key",
		os.Getenv("AZURE_STORAGE_KEY"),
		"Azure storage account key for state backend.\n",
	)
	fl.String(
		"tf-provider-version",
		"",
		"Terraform provider version to use.\n",
	)
	fl.BoolVar(&opts.StrictMode,
		"strict",
		false,
		"Includes cloud provider service-linked roles (disabled by default)",
	)
	fl.BoolVar(&opts.Deep,
		"deep",
		false,
		fmt.Sprintf("%s Enable deep mode\n", warn("EXPERIMENTAL:"))+
			"You should check the documentation for more details: https://docs.driftctl.com/deep-mode\n",
	)
	fl.StringVar(&opts.DriftignorePath,
		"driftignore",
		".driftignore",
		"Path to the driftignore file",
	)
	fl.StringSliceVar(&opts.Driftignores,
		"ignore",
		[]string{},
		fmt.Sprintf("%s Patterns to be used for ignoring resources\n", warn("EXPERIMENTAL:"))+
			"Example: *,!aws_s3* (everything but resources that are prefixed with aws_s3 are ignored) \n"+
			"When using this parameter the driftignore file is not processed\n"+
			"When using multiple instances of this argument, order will be respected")
	fl.String(
		"tf-lockfile",
		".terraform.lock.hcl",
		"Terraform lock file to get the provider's version from. Will be ignored if the file doesn't exist.\n",
	)

	configDir, err := homedir.Dir()
	if err != nil {
		configDir = os.TempDir()
	}
	fl.String(
		"config-dir",
		configDir,
		"Directory path that driftctl uses for configuration.\n",
	)
	fl.BoolVar(&opts.OnlyManaged,
		"only-managed",
		false,
		"Report only what's managed by your IaC\n",
	)
	fl.BoolVar(&opts.OnlyUnmanaged,
		"only-unmanaged",
		false,
		"Report only what's not managed by your IaC\n",
	)

	return cmd
}

func scanRun(opts *pkg.ScanOptions) error {
	store := memstore.New()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	alerter := alerter.NewAlerter()

	// For now, we only use the global printer to print progress and information about the current scan, so unless one
	// of the configured output should silence global output we simply use console by default.
	if output.ShouldPrint(opts.Output, opts.Quiet) {
		globaloutput.ChangePrinter(globaloutput.NewConsolePrinter())
	}

	providerLibrary := terraform.NewProviderLibrary()
	remoteLibrary := common.NewRemoteLibrary()

	iacProgress := globaloutput.NewProgress("Scanning states", "Scanned states", true)
	scanProgress := globaloutput.NewProgress("Scanning resources", "Scanned resources", false)

	resourceSchemaRepository := resource.NewSchemaRepository()

	resFactory := terraform.NewTerraformResourceFactory(resourceSchemaRepository)

	err := remote.Activate(opts.To, opts.ProviderVersion, opts.GCPScope, alerter, providerLibrary, remoteLibrary, scanProgress, resourceSchemaRepository, resFactory, opts.ConfigDir)
	if err != nil {
		return err
	}

	// Teardown
	defer func() {
		logrus.Trace("Exiting scan cmd")
		providerLibrary.Cleanup()
		logrus.Trace("Exited")
	}()

	logrus.Debug("Checking for driftignore")
	driftIgnore := filter.NewDriftIgnore(opts.DriftignorePath, opts.Driftignores...)

	scanner := remote.NewScanner(remoteLibrary, alerter, remote.ScannerOptions{Deep: opts.Deep}, driftIgnore)

	iacSupplier, err := supplier.GetIACSupplier(opts.From, providerLibrary, opts.BackendOptions, iacProgress, alerter, resFactory, driftIgnore)
	if err != nil {
		return err
	}

	ctl := pkg.NewDriftCTL(
		scanner,
		iacSupplier,
		alerter,
		analyser.NewAnalyzer(alerter, analyser.AnalyzerOptions{Deep: opts.Deep, OnlyManaged: opts.OnlyManaged, OnlyUnmanaged: opts.OnlyUnmanaged}, driftIgnore),
		resFactory,
		opts,
		scanProgress,
		iacProgress,
		resourceSchemaRepository,
		store,
	)

	go func() {
		<-c
		logrus.Warn("Detected interrupt, cleanup ...")
		ctl.Stop()
	}()

	analysis, err := ctl.Run()
	if err != nil {
		return err
	}

	analysis.ProviderVersion = resourceSchemaRepository.ProviderVersion.String()
	analysis.ProviderName = resourceSchemaRepository.ProviderName
	store.Bucket(memstore.TelemetryBucket).Set("provider_name", analysis.ProviderName)

	validOutput := false
	for _, o := range opts.Output {
		if err = output.GetOutput(o).Write(analysis); err != nil {
			logrus.Errorf("Error writing to output %s: %v", o.String(), err.Error())
			continue
		}
		validOutput = true
	}

	// Fallback to console output if all output failed
	if !validOutput {
		logrus.Debug("All outputs failed, fallback to console output")
		if err = output.NewConsole().Write(analysis); err != nil {
			return err
		}
	}

	globaloutput.Printf(color.WhiteString("Scan duration: %s\n", analysis.Duration.Round(time.Second)))
	globaloutput.Printf(color.WhiteString("Provider version used to scan: %s. Use --tf-provider-version to use another version.\n"), resourceSchemaRepository.ProviderVersion.String())

	if !opts.DisableTelemetry {
		tl := telemetry.NewTelemetry(&build.Build{})
		tl.SendTelemetry(store.Bucket(memstore.TelemetryBucket))
	}

	if !analysis.IsSync() {
		globaloutput.Printf("\nHint: use gen-driftignore command to generate a .driftignore file based on your drifts\n")

		return cmderrors.InfrastructureNotInSync{}
	}

	return nil
}

func parseScopeFlag(scope []string) ([]string, error) {

	scopeRegex := `projects/\S*$|folders/\d*$|organizations/\d*$`
	r := regexp.MustCompile(scopeRegex)

	for _, v := range scope {
		if !r.MatchString(v) {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					"\nAccepted formats are: projects/<project-id>, folders/<folder-number>, organizations/<org-id>",
				),
				"Unable to parse GCP scope '%s'",
				v,
			)
		}
	}

	return scope, nil
}

func parseFromFlag(from []string) ([]config.SupplierConfig, error) {

	configs := make([]config.SupplierConfig, 0, len(from))

	for _, flag := range from {
		schemePath := strings.Split(flag, "://")
		if len(schemePath) != 2 || schemePath[1] == "" || schemePath[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nAccepted schemes are: %s",
						strings.Join(supplier.GetSupportedSchemes(), ","),
					),
				),
				"Unable to parse from flag '%s'",
				flag,
			)
		}

		scheme := schemePath[0]
		path := schemePath[1]
		supplierBackend := strings.Split(scheme, "+")
		if len(supplierBackend) > 2 {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(fmt.Sprintf(
					"\nAccepted schemes are: %s",
					strings.Join(supplier.GetSupportedSchemes(), ","),
				),
				),
				"Unable to parse from scheme '%s'",
				scheme,
			)
		}

		supplierKey := supplierBackend[0]
		if !supplier.IsSupplierSupported(supplierKey) {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nAccepted values are: %s",
						strings.Join(supplier.GetSupportedSuppliers(), ","),
					),
				),
				"Unsupported IaC source '%s'",
				supplierKey,
			)
		}

		backendString := ""
		if len(supplierBackend) == 2 {
			backendString = supplierBackend[1]
			if !backend.IsSupported(backendString) {
				return nil, errors.Wrapf(
					cmderrors.NewUsageError(
						fmt.Sprintf(
							"\nAccepted values are: %s",
							strings.Join(backend.GetSupportedBackends(), ","),
						),
					),
					"Unsupported IaC backend '%s'",
					backendString,
				)
			}
		}

		configs = append(configs, config.SupplierConfig{
			Key:     supplierKey,
			Backend: backendString,
			Path:    path,
		})
	}

	return configs, nil
}

func parseOutputFlags(out []string) ([]output.OutputConfig, error) {
	result := make([]output.OutputConfig, 0, len(out))
	for _, v := range out {
		o, err := parseOutputFlag(v)
		if err != nil {
			return result, err
		}
		result = append(result, *o)
	}
	return result, nil
}

func parseOutputFlag(out string) (*output.OutputConfig, error) {
	schemeOpts := strings.Split(out, "://")
	if len(schemeOpts) < 2 || schemeOpts[0] == "" {
		return nil, errors.Wrapf(
			cmderrors.NewUsageError(
				fmt.Sprintf(
					"\nAccepted formats are: %s",
					strings.Join(output.SupportedOutputsExample(), ","),
				),
			),
			"Unable to parse output flag '%s'",
			out,
		)
	}

	o := &output.OutputConfig{
		Key: schemeOpts[0],
	}
	if !output.IsSupported(o.Key) {
		return nil, errors.Wrapf(
			cmderrors.NewUsageError(
				fmt.Sprintf(
					"\nValid formats are: %s",
					strings.Join(output.SupportedOutputsExample(), ","),
				),
			),
			"Unsupported output '%s'",
			o.Key,
		)
	}

	opts := schemeOpts[1:]

	switch o.Key {
	case output.JSONOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.JSONOutputType),
					),
				),
				"Invalid json output '%s'",
				out,
			)
		}
		o.Path = opts[0]
	case output.HTMLOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.HTMLOutputType),
					),
				),
				"Invalid html output '%s'",
				out,
			)
		}
		o.Path = opts[0]
	case output.PlanOutputType:
		if len(opts) != 1 || opts[0] == "" {
			return nil, errors.Wrapf(
				cmderrors.NewUsageError(
					fmt.Sprintf(
						"\nMust be of kind: %s",
						output.Example(output.PlanOutputType),
					),
				),
				"Invalid plan output '%s'",
				out,
			)
		}
		o.Path = opts[0]
	}

	return o, nil
}

func validateTfProviderVersionString(version string) error {
	if version == "" {
		return nil
	}
	if match, _ := regexp.MatchString("^\\d+\\.\\d+\\.\\d+$", version); !match {
		return errors.Errorf("Invalid version argument %s, expected a valid semver string (e.g. 2.13.4)", version)
	}
	return nil
}
