package main

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	gosentry "github.com/getsentry/sentry-go"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/build"
	"github.com/snyk/driftctl/logger"
	"github.com/snyk/driftctl/pkg/cmd"
	cmderrors "github.com/snyk/driftctl/pkg/cmd/errors"
	"github.com/snyk/driftctl/pkg/config"
	"github.com/snyk/driftctl/pkg/version"
	"github.com/snyk/driftctl/sentry"
)

// Test
func init() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load() // The Original .env
}

func main() {
	os.Exit(run())
}

func run() int {

	config.Init()
	logger.Init()
	build := build.Build{}
	logrus.WithFields(logrus.Fields{
		"isRelease":               fmt.Sprintf("%t", build.IsRelease()),
		"isUsageReportingEnabled": fmt.Sprintf("%t", build.IsUsageReportingEnabled()),
		"version":                 version.Current(),
	}).Debug("Build info")

	driftctlCmd := cmd.NewDriftctlCmd(build)

	checkVersion := driftctlCmd.ShouldCheckVersion()
	latestVersionChan := make(chan string)
	if checkVersion {
		go func() {
			latestVersion := version.CheckLatest()
			latestVersionChan <- latestVersion
		}()
	}

	// Handle panic and log them to sentry if error reporting is enabled
	defer func() {
		if cmd.IsReportingEnabled(&driftctlCmd.Command) {
			err := recover()
			if err != nil {
				gosentry.CurrentHub().Recover(err)
				flushSentry()
				logrus.Fatalf("Captured panic: %s", err)
				os.Exit(2)
			}
			flushSentry()
		}
	}()

	if _, err := driftctlCmd.ExecuteC(); err != nil {
		if _, isNotInSync := err.(cmderrors.InfrastructureNotInSync); isNotInSync {
			return 1
		}
		if cmd.IsReportingEnabled(&driftctlCmd.Command) {
			sentry.CaptureException(err)
		}
		_, _ = fmt.Fprintln(os.Stderr, color.RedString("%s", err))
		return 1
	}

	if checkVersion {
		newVersion := <-latestVersionChan
		if newVersion != "" {
			_, _ = fmt.Fprintln(os.Stderr, "\n\nYour version of driftctl is outdated, please upgrade !")
			_, _ = fmt.Fprintf(os.Stderr, "Current: %s; Latest: %s\n", version.Current(), newVersion)
		}
	}

	return 0
}

func flushSentry() {
	ttl := 60 * time.Second
	ok := gosentry.Flush(ttl)
	logrus.WithField("timeout", ttl).WithField("success", ok).Debug("Flushed Sentry events")
}
