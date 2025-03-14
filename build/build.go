package build

var env = "dev"

// This flag could be switched to false while building to create a binary without third party network calls
// That mean that following services will be disabled:
// - telemetry
// - version check
var enableUsageReporting = "false"

type BuildInterface interface {
	IsRelease() bool
	IsUsageReportingEnabled() bool
}

type Build struct{}

func (b Build) IsRelease() bool {
	return env == "release"
}

func (b Build) IsUsageReportingEnabled() bool {
	return b.IsRelease() && enableUsageReporting == "true"
}
