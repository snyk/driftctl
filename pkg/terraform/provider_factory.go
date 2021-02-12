package terraform

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
)

func NewGRPCProvider(meta discovery.PluginMeta) (*plugin.GRPCProvider, error) {
	client := Client(meta)
	// Request the RPC terraformProvider so we can get the provider
	// so we can build the actual RPC-implemented provider.
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}
	raw, err := rpcClient.Dispense(plugin.ProviderPluginName)
	if err != nil {
		return nil, err
	}

	GRPCProvider := raw.(*plugin.GRPCProvider)
	GRPCProvider.PluginClient = client

	return GRPCProvider, nil
}
