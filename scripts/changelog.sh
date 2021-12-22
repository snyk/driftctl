#!/usr/bin/env bash

# This script compare merged pull requests between the two most recent tags
# Please note that this script only work with Github repositories.
# Prerequisites: git, github cli

GHCLI_BIN="gh"
REPO="snyk/driftctl"
LATEST_TAG=$(git describe --abbrev=0) # Get the last created tag
DEFAULT_BRANCH=origin/HEAD
BASE=$(git for-each-ref --sort=-taggerdate --format '%(tag)' refs/tags | sed -n 2p) # Use $DEFAULT_BRANCH instead to get a pre-release changelog

# Check GH cli is installed
if ! which $GHCLI_BIN &> /dev/null; then
    echo "GitHub CLI ($GHCLI_BIN) is not installed, visit https://github.com/cli/cli#installation"
    exit 1
fi

# Check GH authentication
if ! $GHCLI_BIN auth status &> /dev/null; then
    echo "You are not logged into any GitHub hosts. Run gh auth login to authenticate."
    exit 1
fi

# Compare $BASE branch with the latest tag
# Keep IDs of merged pull requests
PRs=$(git log --pretty=oneline $BASE...$LATEST_TAG | grep 'Merge pull request #' | grep -oP '#[0-9]+' | sed 's/#//')

# Generating changelog for commits from $BASE to $LATEST_TAG
CHANGES=()
for pr in $PRs; do
    str=$($GHCLI_BIN pr view $pr --repo $REPO -t '- {{ .title }} (#{{ .number }}) @{{ .author.login }} {{.labels}}' --json title,number,author,labels)
    CHANGES+=("$str")
done

print_changes() {
    local label=$1
    local title=$2
    if [[ "${CHANGES[@]}" =~ $label ]]; then
        echo -e $title
        for change in "${CHANGES[@]}"; do
            if [[ $change =~ $label ]]; then
                echo $change | sed "s/\[map\[$PARTITION_COLUMN.*//"
            fi
        done
    fi
}

print_changes "kind/enhancement" "## 🚀 Enhancements"
print_changes "kind/bug" "## 🐛 Bug Fixes"
print_changes "kind/maintenance" "## 🔨 Maintenance"
