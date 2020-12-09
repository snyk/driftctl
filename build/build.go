package build

var env = "dev"

type BuildInterface interface {
	IsRelease() bool
}

type Build struct{}

func (b Build) IsRelease() bool {
	return env == "release"
}
