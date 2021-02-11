package terraform

import (
	"os/exec"

	"github.com/cloudskiff/driftctl/logger"
	tfplugin "github.com/hashicorp/terraform/plugin"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
)

func ClientConfig(m discovery.PluginMeta) *plugin.ClientConfig {
	logger := logger.NewTerraformPluginLogger()
	return &plugin.ClientConfig{
		Cmd:              exec.Command(m.Path),
		HandshakeConfig:  tfplugin.Handshake,
		VersionedPlugins: tfplugin.VersionedPlugins,
		Managed:          true,
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         true,
	}
}

// Client returns a plugin client for the plugin described by the given metadata.
func Client(m discovery.PluginMeta) *plugin.Client {
	return plugin.NewClient(ClientConfig(m))
}
