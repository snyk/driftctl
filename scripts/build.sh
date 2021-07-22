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

# Check configuration
goreleaser check

# Build!
echo "+ Building using goreleaser ..."
ENV=dev goreleaser build \
    --rm-dist \
    --parallelism 2 \
    --snapshot \
    --single-target
