package version

import (
	"encoding/json"
	"io"
	"net/http"
	"runtime"

	"github.com/snyk/driftctl/build"

	goversion "github.com/hashicorp/go-version"
	"github.com/sirupsen/logrus"
)

// Current software version
// Could be injected on build with -ldflags
var version string = "dev"

// Return the current version as string
func Current() string {
	currentVersion := version
	build := build.Build{}
	if !build.IsRelease() {
		currentVersion += "-dev"
	}
	return currentVersion
}

/**
 * Return "" if current version is the last version,
 * else return the latest version string
 **/
func CheckLatest() string {

	client := &http.Client{}

	req, _ := http.NewRequest("GET", "https://telemetry.driftctl.com/version", nil)
	req.Header.Set("driftctl-version", Current())
	req.Header.Set("driftctl-os", runtime.GOOS)
	req.Header.Set("driftctl-arch", runtime.GOARCH)

	res, err := client.Do(req)
	if err != nil {
		logrus.Debugf("Unable to check for a newer version: %+v", err)
		return ""
	}

	if res.StatusCode != 200 {
		logrus.Debugf("Unable to check for a newer version: %s", res.Status)
		return ""
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Debug("Unable to read response")
		logrus.Debug(err)
		return ""
	}

	responseBody := map[string]string{}
	err = json.Unmarshal(bodyBytes, &responseBody)
	if err != nil {
		logrus.Debug("Unable to decode version check response")
		logrus.Debug(err)
		return ""
	}

	currentVersion, err := goversion.NewVersion(version)
	if err != nil {
		logrus.Debugf("Unable to parse current version: %s", version)
		return ""
	}

	lastVersion, err := goversion.NewVersion(responseBody["latest"])
	if err != nil {
		logrus.Debugf("Unable to parse latest version: %s", responseBody["latest"])
		return ""
	}

	if currentVersion.LessThan(lastVersion) {
		return lastVersion.String()
	}

	return ""
}
