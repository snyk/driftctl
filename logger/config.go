package logger

import (
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/build"
	"github.com/spf13/viper"
)

func getConfig() Config {

	config := Config{
		Level:        logrus.WarnLevel,
		ReportCaller: false,
		Formatter:    NewTextFormatter(4),
	}

	build := build.Build{}
	if !build.IsRelease() {
		config.Level = logrus.DebugLevel
	}

	if viper.IsSet("log_level") {
		level, _ := logrus.ParseLevel(viper.GetString("log_level"))
		config.Level = level
	}

	return config
}
