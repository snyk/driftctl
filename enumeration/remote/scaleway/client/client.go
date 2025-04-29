package client

import (
	"errors"
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/snyk/driftctl/pkg/version"
)

// Create - Creates a new Scaleway client object
// Heavily inspired by the createClient method in Scaleway CLI:
// https://github.com/scaleway/scaleway-cli/blob/v2.13.0/internal/core/client.go#L15-L83
func Create() (*scw.Client, error) {

	profile := scw.LoadEnvProfile()

	// Default path is based on the following priority order:
	// * $SCW_CONFIG_PATH
	// * $XDG_CONFIG_HOME/scw/config.yaml
	// * $HOME/.config/scw/config.yaml
	// * $USERPROFILE/.config/scw/config.yaml
	var errConfigFileNotFound *scw.ConfigFileNotFoundError
	config, err := scw.LoadConfigFromPath(scw.GetConfigPath())

	switch {
	case errors.As(err, &errConfigFileNotFound):
		break
	case err != nil:
		return nil, err
	default:
		// If a config file is found and loaded, we merge with env
		activeProfile, err := config.GetActiveProfile()
		if err != nil {
			return nil, err
		}

		// Creates a client from the active profile
		// It will trigger a validation step on its configuration to catch errors if any
		opts := []scw.ClientOption{
			scw.WithProfile(activeProfile),
		}

		_, err = scw.NewClient(opts...)
		if err != nil {
			return nil, err
		}

		profile = scw.MergeProfiles(activeProfile, profile)
	}

	opts := []scw.ClientOption{
		scw.WithDefaultRegion(scw.RegionFrPar),
		scw.WithDefaultZone(scw.ZoneFrPar1),
		scw.WithUserAgent(fmt.Sprintf("driftctl/%s", version.Current())),
		scw.WithProfile(profile),
	}

	client, err := scw.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
}
