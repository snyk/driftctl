package terraform

import (
	"fmt"
	"runtime"
)

type ProviderConfig struct {
	Key       string
	Version   string
	ConfigDir string
}

func (c *ProviderConfig) GetDownloadUrl() string {
	arch := runtime.GOARCH
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		arch = "amd64"
	}
	return fmt.Sprintf(
		"https://releases.hashicorp.com/terraform-provider-%s/%s/terraform-provider-%s_%s_%s_%s.zip",
		c.Key,
		c.Version,
		c.Key,
		c.Version,
		runtime.GOOS,
		arch,
	)
}

func (c *ProviderConfig) GetBinaryName() string {
	return fmt.Sprintf("terraform-provider-%s_v%s", c.Key, c.Version)
}
