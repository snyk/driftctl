package mocks

type MockBuild struct {
	Release        bool
	UsageReporting bool
}

func (m MockBuild) IsRelease() bool {
	return m.Release
}

func (m MockBuild) IsUsageReportingEnabled() bool {
	return m.UsageReporting
}
