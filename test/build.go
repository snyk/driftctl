package test

type Build struct{}

func (b Build) IsRelease() bool {
	return false
}

func (b Build) IsUsageReportingEnabled() bool {
	return false
}
