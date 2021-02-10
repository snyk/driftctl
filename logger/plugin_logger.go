package logger

import (
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

type terraformPluginFormatter struct {
	logrus.Formatter
}

func (f *terraformPluginFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Message = "[TerraformPlugin] " + entry.Message
	return f.Formatter.Format(entry)
}

var levelMap = map[hclog.Level]logrus.Level{
	hclog.Trace: logrus.TraceLevel,
	hclog.Debug: logrus.DebugLevel,
	hclog.Info:  logrus.InfoLevel,
	hclog.Warn:  logrus.WarnLevel,
	hclog.Error: logrus.ErrorLevel,
}

func resolveHCLLogLevels(level hclog.Level) logrus.Level {
	return levelMap[level]
}

type TerraformPluginLogger struct {
	logger *logrus.Logger
}

func NewTerraformPluginLogger() TerraformPluginLogger {
	config := getConfig()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	logger.SetReportCaller(false)
	logger.SetFormatter(&terraformPluginFormatter{Formatter: config.Formatter})

	// Disable terraform provider log if we are not in trace level
	if config.Level == logrus.TraceLevel {
		logger.SetLevel(logrus.TraceLevel)
	}

	return TerraformPluginLogger{logger}
}

func (t TerraformPluginLogger) Trace(msg string, args ...interface{}) {
	t.logger.Trace(msg, args)
}

func (t TerraformPluginLogger) Debug(msg string, args ...interface{}) {
	t.logger.Debug(msg, args)
}

func (t TerraformPluginLogger) Info(msg string, args ...interface{}) {
	t.logger.Info(msg, args)
}

func (t TerraformPluginLogger) Warn(msg string, args ...interface{}) {
	t.logger.Warn(msg, args)
}

func (t TerraformPluginLogger) Error(msg string, args ...interface{}) {
	t.logger.Error(msg, args)
}

func (t TerraformPluginLogger) IsTrace() bool {
	return t.logger.IsLevelEnabled(logrus.TraceLevel)
}

func (t TerraformPluginLogger) IsDebug() bool {
	return t.logger.IsLevelEnabled(logrus.DebugLevel)
}

func (t TerraformPluginLogger) IsInfo() bool {
	return t.logger.IsLevelEnabled(logrus.InfoLevel)
}

func (t TerraformPluginLogger) IsWarn() bool {
	return t.logger.IsLevelEnabled(logrus.WarnLevel)
}

func (t TerraformPluginLogger) IsError() bool {
	return t.logger.IsLevelEnabled(logrus.ErrorLevel)
}

func (t TerraformPluginLogger) With(args ...interface{}) hclog.Logger {
	return t
}

func (t TerraformPluginLogger) Named(name string) hclog.Logger {
	return t
}

func (t TerraformPluginLogger) ResetNamed(name string) hclog.Logger {
	return t
}

func (t TerraformPluginLogger) SetLevel(level hclog.Level) {
	t.logger.SetLevel(resolveHCLLogLevels(level))
}

func (t TerraformPluginLogger) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	stdLogger := log.New(t.logger.Writer(), "", log.Flags())
	stdLogger.SetOutput(t.logger.Writer())
	return stdLogger
}

func (t TerraformPluginLogger) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return t.logger.Writer()
}

func (t TerraformPluginLogger) Log(level hclog.Level, msg string, args ...interface{}) {
	t.logger.Log(resolveHCLLogLevels(level), msg, args)
}

func (t TerraformPluginLogger) ImpliedArgs() []interface{} {
	return nil
}

func (t TerraformPluginLogger) Name() string {
	return "TerraformPlugin"
}
