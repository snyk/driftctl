package mocks

type MockBuild struct {
	Release bool
}

func (m MockBuild) IsRelease() bool {
	return m.Release
}
