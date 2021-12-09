#!/usr/bin/env bash

# Please note that this script only work with Github repositories.
# Prerequisites: git, github cli

GHCLI_BIN="gh"
REPO="snyk/driftctl"
LATEST_TAG=$(git describe --abbrev=0) # Get the least created tag
DEFAULT_BRANCH=origin/HEAD
BASE=$DEFAULT_BRANCH # Change this if you don't want to use the default branch as base

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

echo "Generating changelog for commits from $BASE to $LATEST_TAG..."

CHANGES=()
for pr in $PRs; do
    str=$($GHCLI_BIN pr view $pr --repo $REPO -t '- {{ .title }} (#{{ .number }}) @{{ .author.login }} {{.labels}}' --json title,number,author,labels)
    CHANGES+=("$str")
done

echo -e "\n## 🚀 Enhancements\n"

for change in "${CHANGES[@]}"; do
    if [[ $change =~ "kind/enhancement" ]]; then
        echo $change | sed "s/\[map\[$PARTITION_COLUMN.*//"
    fi
done

echo -e "\n## 🐛 Bug Fixes\n"

for change in "${CHANGES[@]}"; do
    if [[ $change =~ "kind/bug" ]]; then
        echo $change | sed "s/\[map\[$PARTITION_COLUMN.*//"
    fi
done

echo -e "\n## 🔨 Maintenance\n"

for change in "${CHANGES[@]}"; do
    if [[ $change =~ "kind/maintenance" ]]; then
        echo $change | sed "s/\[map\[$PARTITION_COLUMN.*//"
    fi
done
