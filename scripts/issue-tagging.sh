#!/usr/bin/env bash
# This script compares merged pull requests between the two most recent tags
# Please note that this script only works with Github repositories.
# Prerequisites: git, github cli, curl

set -euo pipefail

GHCLI_BIN="gh"
REPO="snyk/driftctl"
LATEST_TAG=$(git for-each-ref --sort=-taggerdate --format '%(tag)' refs/tags | sed -n 1p) # Get the latest created tag
BASE_TAG=$(git for-each-ref --sort=-taggerdate --format '%(tag)' refs/tags | sed -n 2p) # Get the second latest created tag

# Check GH CLI is installed
if ! which $GHCLI_BIN &> /dev/null; then
    echo "GitHub CLI ($GHCLI_BIN) is not installed, visit https://github.com/cli/cli#installation"
    exit 1
fi

# Check GH authentication
if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "GITHUB_TOKEN environment variable is not set, it is required to use the GitHub API."
    exit 1
fi

# Check GH authentication
if ! $GHCLI_BIN auth status &> /dev/null; then
    echo "You are not logged into any GitHub hosts. Run gh auth login to authenticate."
    exit 1
fi

# Compare $BASE_TAG branch with the latest tag
# Keep IDs of merged pull requests
PRs=$(git log --pretty=oneline "$BASE_TAG"..."$LATEST_TAG" | grep 'Merge pull request #' | grep -oE '#[0-9]+' | sed 's/#//')

# Find fixed issues from $BASE_TAG to $LATEST_TAG
ISSUES=()
for pr in $PRs; do
    id=$($GHCLI_BIN pr view "$pr" --json body | grep -oE 'Related issues | #[0-9]+' | sed 's/[^[:digit:]]//g' | sed -z 's/\n//g')
    ISSUES+=("$id")
done

echo "Creating milestone $BASE_TAG in github.com/$REPO"
curl -X POST \
    -H "Accept: application/vnd.github.v3+json" \
    -H "Authorization: token $GITHUB_TOKEN" \
    --data "{\"title\":\"$BASE_TAG\"}" \
    "https://api.github.com/repos/$REPO/milestone"

for issue in "${ISSUES[@]}"; do
    if [ -z "$issue" ]; then
        continue
    fi
    echo "Adding milestone $BASE_TAG to issue #$issue"
    gh issue edit "$issue" -m "$BASE_TAG"

    curl -X POST \
        -H "Accept: application/vnd.github.v3+json" \
        -H "Authorization: token $GITHUB_TOKEN" \
        --data "{\"body\":\"This issue has been referenced in the latest release $BASE_TAG.\"}" \
        "https://api.github.com/repos/$REPO/issues/$issue/comments"
done

echo "Done."
