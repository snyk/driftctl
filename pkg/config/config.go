package config

import "github.com/spf13/viper"

func Init() {
	_ = viper.BindEnv("log_level")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("dctl")
}

func IsSnyk() bool {
	return viper.GetBool("IS_SNYK")
}
