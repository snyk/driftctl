package cmd

import (
	"encoding/json"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/snyk/driftctl/pkg/alerter"
	globaloutput "github.com/snyk/driftctl/pkg/output"
	"github.com/snyk/driftctl/pkg/remote"
	"github.com/snyk/driftctl/pkg/remote/common"
	"github.com/snyk/driftctl/pkg/resource"
	"github.com/snyk/driftctl/pkg/terraform"
	"github.com/spf13/cobra"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud resources.",
		Long:  "List cloud resources by type.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resourceType := args[0]

			providerLibrary := terraform.NewProviderLibrary()
			remoteLibrary := common.NewRemoteLibrary()
			alerter := alerter.NewAlerter()
			scanProgress := globaloutput.NewProgress("Scanning resources", "Scanned resources", false)
			resourceSchemaRepository := resource.NewSchemaRepository()
			resFactory := terraform.NewTerraformResourceFactory(resourceSchemaRepository)

			home, err := homedir.Dir()
			must(err)

			// hardcoded some AWS parameters for spike purposes only
			err = remote.Activate(
				"aws+tf",
				"",
				alerter, providerLibrary,
				remoteLibrary, scanProgress,
				resourceSchemaRepository,
				resFactory, home,
			)
			must(err)

			scanner := remote.NewScanner(remoteLibrary, alerter, remote.ScannerOptions{Deep: true}, nil)
			resources, err := scanner.Resources()
			must(err)

			var filteredResources []*resource.Resource
			for _, resource := range resources {
				if resource.Type == resourceType || resourceType == "all" {
					filteredResources = append(filteredResources, resource)
				}
			}
			must(json.NewEncoder(os.Stdout).Encode(filteredResources))
		},
	}
	return cmd
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
