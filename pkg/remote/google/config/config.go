package config

type GCPTerraformConfig struct {
	Project string `cty:"project"`
	Region  string `cty:"region"`
	Zone    string `cty:"zone"`
}
