package terraform

import (
	"fmt"
	"runtime"
)

type ProviderConfig struct {
	Key     string
	Version string
	Postfix string
}

func (c *ProviderConfig) GetDownloadUrl() string {
	arch := runtime.GOOS
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
	if c.Postfix == "" {
		return fmt.Sprintf("terraform-provider-%s_v%s", c.Key, c.Version)
	}
	return fmt.Sprintf("terraform-provider-%s_v%s_%s", c.Key, c.Version, c.Postfix)
}
