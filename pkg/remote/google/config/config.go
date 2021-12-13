package config

type GCPTerraformConfig struct {
	Organization string `cty:"organization"`
	Project      string `cty:"project"`
	Region       string `cty:"region"`
	Zone         string `cty:"zone"`
}
