package main

import (
	"fmt"
	"os"

	"github.com/cloudskiff/driftctl/build"
	"github.com/cloudskiff/driftctl/logger"
	"github.com/cloudskiff/driftctl/pkg/cmd"
	"github.com/cloudskiff/driftctl/pkg/config"
	"github.com/cloudskiff/driftctl/pkg/version"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
)

func init() {
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load() // The Original .env
}

func main() {

	config.Init()
	logger.Init(logger.GetConfig())

	driftctlCmd := cmd.NewDriftctlCmd(build.Build{})

	checkVersion := driftctlCmd.ShouldCheckVersion()
	latestVersionChan := make(chan string)
	if checkVersion {
		go func() {
			latestVersion := version.CheckLatest()
			latestVersionChan <- latestVersion
		}()
	}

	if _, err := driftctlCmd.ExecuteC(); err != nil {
		fmt.Fprintln(os.Stderr, color.RedString("%s", err))
		os.Exit(1)
	}

	if checkVersion {
		newVersion := <-latestVersionChan
		if newVersion != "" {
			fmt.Println("\n\nYour version of driftctl is outdated, please upgrade !")
			fmt.Printf("Current: %s; Latest: %s\n", version.Current(), newVersion)
		}
	}
}
