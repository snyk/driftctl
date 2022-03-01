#!/usr/bin/env bash

# This script compare merged pull requests between the two most recent tags
# Please note that this script only work with Github repositories.
# Prerequisites: git, jq, github cli

print_changelist() {
  title=$1
  shift
  list=("$@")

  echo -e "$title"
  for change in "${list[@]}"; do
    echo "$change"
  done
}

GHCLI_BIN="gh"
JQ_BIN="jq"
REPO="snyk/driftctl"
LATEST_TAG=$(git for-each-ref --sort=-taggerdate --format '%(tag)' refs/tags | sed -n 1p) # Get the last created tag
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
BASE=$(git for-each-ref --sort=-taggerdate --format '%(tag)' refs/tags | sed -n 2p) # Use $CURRENT_BRANCH instead to get a pre-release changelog

# Check GH cli is installed
if ! which $GHCLI_BIN &> /dev/null; then
    echo "GitHub CLI ($GHCLI_BIN) is not installed, visit https://github.com/cli/cli#installation"
    exit 1
fi

# Check jq is installed
if ! which $JQ_BIN &> /dev/null; then
    echo "jq ($JQ_BIN) is not installed"
    exit 1
fi

# Check GH authentication
if ! $GHCLI_BIN auth status &> /dev/null; then
    echo "You are not logged into any GitHub hosts. Run gh auth login to authenticate."
    exit 1
fi

# Compare $BASE branch with the latest tag
# Keep IDs of merged pull requests
PRs=$(git log --pretty=oneline "$BASE"..."$LATEST_TAG" | grep 'Merge pull request #' | grep -oP '#[0-9]+' | sed 's/#//')

# Generating changelog for commits from $BASE to $LATEST_TAG
enchancements=()
fixes=()
maintenance=()
uncategorised=()

for pr in $PRs; do
    json=$($GHCLI_BIN pr view "$pr" --repo $REPO --json title,number,author,labels)

    labels=$(echo "$json" | jq .labels[].name)
    title=$(echo "$json" | jq -r .title)
    number=$(echo "$json" | jq -r .number)
    author=$(echo "$json" | jq -r .author.login)

    str="- $title (#$number) @$author"

    if [[ $labels =~ "kind/enhancement" ]]; then
        enchancements+=("$str")
    elif [[ $labels =~ "kind/bug" ]]; then
        fixes+=("$str")
    elif [[ $labels =~ "kind/maintenance" ]]; then
        maintenance+=("$str")
    else
        uncategorised+=("$str")
    fi
done

print_changelist "## 🚀 Enhancements" "${enchancements[@]}"
print_changelist "## 🐛 Bug Fixes" "${fixes[@]}"
print_changelist "## 🔨 Maintenance" "${maintenance[@]}"
print_changelist "## Uncategorised" "${uncategorised[@]}"
