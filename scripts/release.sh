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

GRFLAGS=""

# We sign every releases using PGP
# We may not want to do so in dev environments
if [ -z $SIGNINGKEY ]; then
    GRFLAGS="--skip-sign"
fi

# Only CI system should publish artifacts
if [ "$CI" != "circleci" ]; then
    GRFLAGS="--snapshot"
    GRFLAGS="--skip-announce"
    GRFLAGS="--skip-publish"
fi

# Build!
echo "+ Building using goreleaser ..."
ENV=release goreleaser release \
    --rm-dist \
    --parallelism 2 \
    ${GRFLAGS}
