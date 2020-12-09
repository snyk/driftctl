package test

type Build struct{}

func (b Build) IsRelease() bool {
	return false
}
