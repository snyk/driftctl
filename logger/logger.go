package logger

import (
	"io"
	"log"

	"github.com/sirupsen/logrus"
)

type Config struct {
	Level        logrus.Level
	Formatter    logrus.Formatter
	ReportCaller bool
}

func Init(loggerConfig Config) {
	logrus.SetLevel(loggerConfig.Level)
	logrus.SetReportCaller(loggerConfig.ReportCaller)
	logrus.SetFormatter(loggerConfig.Formatter)

	// Libs that use logger (like grpc provider) will log at TRACE level
	log.SetOutput(GetTraceWriter())
}

// Get a writer which will log at trace level
func GetTraceWriter() io.Writer {
	redirectLogger := logrus.New()
	return redirectLogger.WriterLevel(logrus.TraceLevel)
}
