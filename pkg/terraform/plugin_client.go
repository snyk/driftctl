package terraform

import (
	"github.com/cloudskiff/driftctl/logger"
	"os/exec"

	tfplugin "github.com/hashicorp/terraform/plugin"

	"github.com/hashicorp/go-hclog"

	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform/plugin/discovery"
)

func ClientConfig(m discovery.PluginMeta) *plugin.ClientConfig {
	// redirect plugin logger to trace level in logrus
	pluginLogger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Level:  hclog.Trace,
		Output: logger.GetTraceWriter(),
	})

	return &plugin.ClientConfig{
		Cmd:              exec.Command(m.Path),
		HandshakeConfig:  tfplugin.Handshake,
		VersionedPlugins: tfplugin.VersionedPlugins,
		Managed:          true,
		Logger:           pluginLogger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		AutoMTLS:         true,
	}
}

// Client returns a plugin client for the plugin described by the given metadata.
func Client(m discovery.PluginMeta) *plugin.Client {
	return plugin.NewClient(ClientConfig(m))
}
