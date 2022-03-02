package config

type GCPTerraformConfig struct {
	Scopes []string `cty:"scopes"`
}
