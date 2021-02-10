package logger

import (
	"log"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Level        logrus.Level
	Formatter    logrus.Formatter
	ReportCaller bool
}

func Init() {
	config := getConfig()
	logrus.SetLevel(config.Level)
	logrus.SetReportCaller(config.ReportCaller)
	logrus.SetFormatter(config.Formatter)

	// Libs that use logger (like grpc provider) will log at TRACE level
	redirectLogger := logrus.New()
	redirectLogger.SetLevel(config.Level)
	redirectLogger.SetFormatter(config.Formatter)
	log.SetOutput(redirectLogger.WriterLevel(logrus.TraceLevel))
}
