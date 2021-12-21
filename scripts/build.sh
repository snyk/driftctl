#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
# Inspired from hashicorp/terraform build script
# https://github.com/hashicorp/terraform/blob/83e6703bf77f60660db4465ef50d30c633f800f1/scripts/build.sh
set -eo pipefail

if ! which goreleaser >/dev/null; then
    echo "+ Installing goreleaser..."
    go install github.com/goreleaser/goreleaser@v1.1.0
fi

export ENV="${ENV:-dev}"
SINGLE_TARGET="${SINGLE_TARGET:-false}"

# Check configuration
goreleaser check

FLAGS=""
FLAGS+="--rm-dist "
FLAGS+="--parallelism 2 "

CMD="release"

if [ "$SINGLE_TARGET" == "true" ]; then
    CMD="build"
    FLAGS+="--single-target "
fi

# Only CI system should publish artifacts
# We may not want to sign artifacts in dev environments
if [ "$CI" != true ] && [ "$CMD" == "release" ]; then
    FLAGS+="--skip-announce "
    FLAGS+="--skip-publish "
    FLAGS+="--skip-sign "
fi

if [ "$CI" != true ]; then
    FLAGS+="--snapshot "
fi

if [ "$CI" == true ] && [ "$CMD" == "release" ]; then
    echo "Generating changelog..."
    ./scripts/changelog.sh > CHANGELOG.md
    cat CHANGELOG.md
    FLAGS+="--release-notes CHANGELOG.md "
fi

CMD="goreleaser ${CMD} ${FLAGS}"

echo "+ Building using goreleaser"
echo "+ ENV=${ENV}"
echo "+ CMD=${CMD}"

$CMD
