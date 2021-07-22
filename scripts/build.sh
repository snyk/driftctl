#!/usr/bin/env bash
#
# This script builds the application from source for multiple platforms.
# Inspired from hashicorp/terraform build script
# https://github.com/hashicorp/terraform/blob/83e6703bf77f60660db4465ef50d30c633f800f1/scripts/build.sh
set -eo pipefail

if ! which goreleaser >/dev/null; then
    echo "+ Installing goreleaser..."
    go install github.com/goreleaser/goreleaser@v0.173.2
fi

if [ -z $ENV ]; then
    echo "Error: ENV variable must be defined"
    exit 1
fi

# Check configuration
goreleaser check

if [ "$ENV" == "dev" ]; then
    echo "+ Building using goreleaser ..."
    goreleaser build \
        --rm-dist \
        --parallelism 2 \
        --snapshot \
        --single-target
    exit 0
fi

GRFLAGS=""

# Only CI system should publish artifacts
# We may not want to sign artifacts in dev environments
if [ "$CI" != true ]; then
    GRFLAGS+="--snapshot "
    GRFLAGS+="--skip-announce "
    GRFLAGS+="--skip-publish "
    GRFLAGS+="--skip-sign "
fi

echo ${GRFLAGS}

echo "+ Building using goreleaser ..."
goreleaser release \
    --rm-dist \
    ${GRFLAGS}
