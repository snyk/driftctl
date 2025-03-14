package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/build"
	"github.com/snyk/driftctl/enumeration/alerter"
	"github.com/snyk/driftctl/enumeration/remote"
	"github.com/snyk/driftctl/enumeration/remote/aws"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/terraform"
	"github.com/snyk/driftctl/enumeration/terraform/lock"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/pkg/iac/config"
	"github.com/snyk/driftctl/pkg/iac/terraform/state"
	"github.com/snyk/driftctl/pkg/memstore"
	dctlresource "github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/resource/schemas"
	"github.com/snyk/driftctl/pkg/telemetry"
	"github.com/snyk/driftctl/pkg/terraform/hcl"
	"github.com/spf13/cobra"

	"github.com/snyk/driftctl/pkg"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"
	"github.com/snyk/driftctl/pkg/cmd/scan/output"
	"github.com/snyk/driftctl/pkg/filter"
	"github.com/snyk/driftctl/pkg/iac/supplier"
	"github.com/snyk/driftctl/pkg/iac/terraform/state/backend"
	globaloutput "github.com/snyk/driftctl/pkg/output"
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
		[]string{},
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
	var deprecatedOnlyUnmanaged bool
	fl.BoolVar(&deprecatedOnlyUnmanaged,
		"only-unmanaged",
		false,
		fmt.Sprintf("%s Report only what's not managed by your IaC.\nThis option is a no-op as unmanaged is the only supported mode.\n", warn("DEPRECATED:")),
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

	if len(opts.From) == 0 {
		supplierConfigs, err := retrieveBackendsFromHCL("")
		if err != nil {
			return err
		}
		opts.From = append(opts.From, supplierConfigs...)
	}

	if len(opts.From) == 0 {
		opts.From = append(opts.From, config.SupplierConfig{
			Key:     state.TerraformStateReaderSupplier,
			Backend: backend.BackendKeyFile,
			Path:    "terraform.tfstate",
		})
	}

	providerLibrary := terraform.NewProviderLibrary()
	remoteLibrary := common.NewRemoteLibrary()

	iacProgress := globaloutput.NewProgress("Scanning states", "Scanned states", true)
	scanProgress := globaloutput.NewProgress("Scanning resources", "Scanned resources", false)

	resourceSchemaRepository := schemas.NewSchemaRepository()

	resFactory := dctlresource.NewDriftctlResourceFactory(resourceSchemaRepository)

	err := remote.Activate(opts.To, opts.ProviderVersion, alerter, providerLibrary, remoteLibrary, scanProgress, resFactory, opts.ConfigDir)
	if err != nil {
		if err == aws.AWSCredentialsNotFoundError {
			// special case command-line advice, because AWS is the default cloud
			// provider, and users may be confused by a cloud-specific error out of
			// the box
			return fmt.Errorf("%s\n\n%s", err, "To use a different cloud provider, use --to=\"gcp+tf\" for GCP or --to=\"azure+tf\" for Azure.")
		}
		return err
	}

	providerName := common.RemoteParameter(opts.To).GetProviderAddress().Type
	err = resourceSchemaRepository.Init(providerName, opts.ProviderVersion, providerLibrary.Provider(providerName).Schema())
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

	// Create a composite filter that includes both driftIgnore and opts.Filter
	compositeFilter := filter.NewCompositeFilter(driftIgnore, opts.Filter)
	scanner := remote.NewScanner(remoteLibrary, alerter, compositeFilter)

	iacSupplier, err := supplier.GetIACSupplier(opts.From, providerLibrary, opts.BackendOptions, iacProgress, alerter, resFactory, driftIgnore)
	if err != nil {
		return err
	}

	ctl := pkg.NewDriftCTL(
		scanner,
		iacSupplier,
		alerter,
		analyser.NewAnalyzer(alerter, driftIgnore),
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

	analysis.ProviderVersion = opts.ProviderVersion
	analysis.ProviderName = opts.To
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
	globaloutput.Printf(color.WhiteString("Provider version used to scan: %s. Use --tf-provider-version to use another version.\n"), opts.ProviderVersion)

	if !opts.DisableTelemetry {
		tl := telemetry.NewTelemetry(&build.Build{})
		tl.SendTelemetry(store.Bucket(memstore.TelemetryBucket))
	}

	if !analysis.IsSync() {
		return cmderrors.InfrastructureNotInSync{}
	}

	return nil
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

func retrieveBackendsFromHCL(workdir string) ([]config.SupplierConfig, error) {
	matches, err := filepath.Glob(path.Join(workdir, "*.tf"))
	if err != nil {
		return nil, err
	}
	supplierConfigs := make([]config.SupplierConfig, 0)

	for _, match := range matches {
		body, err := hcl.ParseTerraformFromHCL(match)
		if err != nil {
			logrus.
				WithField("file", match).
				WithField("error", err).
				Debug("Error parsing backend block in Terraform file")
			continue
		}

		var cfg *config.SupplierConfig
		ws := hcl.GetCurrentWorkspaceName(path.Dir(match))

		if body.Cloud != nil {
			cfg = body.Cloud.SupplierConfig(ws)
		}
		if body.Backend != nil {
			cfg = body.Backend.SupplierConfig(ws)
		}
		if cfg != nil {
			globaloutput.Printf(color.WhiteString("Using Terraform state %s found in %s. Use the --from flag to specify another state file.\n"), cfg, match)
			supplierConfigs = append(supplierConfigs, *cfg)
		}
	}

	return supplierConfigs, nil
}
