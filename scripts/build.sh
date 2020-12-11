
#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
# Inspired from hashicorp/terraform build script
set -euo pipefail

echo "Bash: ${BASH_VERSION}"

# By default build for dev
ENV=${ENV:-"dev"}

# Get the git commit
GIT_COMMIT=$(git rev-parse HEAD)
GIT_DIRTY=$(test -n "`git status --porcelain`" && echo "-dev" || true)

# Retrieve
VERSION=$(git describe --tags 2>/dev/null || git rev-parse --short HEAD)

# Inject version number
LD_FLAGS="-X github.com/cloudskiff/driftctl/pkg/version.version=${VERSION}"

# Reference:
# https://github.com/golang/go/blob/master/src/go/build/syslist.go
os_archs=(
    darwin/amd64
    linux/386
    linux/amd64
    linux/arm
    linux/arm64
    windows/386
    windows/amd64
)

if [ -n "$OS_ARCH" ]; then
  os_archs=("$OS_ARCH")
fi

echo "ARCH: $os_archs"

if [ $ENV != "release" ]; then
    echo "+ Building env: dev"
    # If its dev mode, only build for ourself
    os_archs=("$(go env GOOS)/$(go env GOARCH)")
    # And set version to git commit
    VERSION="${GIT_COMMIT}${GIT_DIRTY}"
fi

# In release mode we don't want debug information in the binary
# We also set the build env to release
if [ $ENV = "release" ]; then
    echo "+ Building env: release"
    LD_FLAGS="-s -w -X github.com/cloudskiff/driftctl/build.env=release ${LD_FLAGS}"
fi

if ! which gox > /dev/null; then
    echo "+ Installing gox..."
    go get github.com/mitchellh/gox
fi

# Delete old binaries
echo "+ Removing old binaries ..."
rm -f bin/*

# Instruct gox to build statically linked binaries
export CGO_ENABLED=0

# Ensure all remote modules are downloaded and cached before build so that
# the concurrent builds launched by gox won't race to redundantly download them.
go mod download

# Build!
echo "+ Building with flags: ${LD_FLAGS}"
osarch="${os_archs[@]}"
gox \
    -osarch="$osarch" \
    -ldflags "${LD_FLAGS}" \
    -output "bin/driftctl_{{.OS}}_{{.Arch}}" \
    ./
