#!/usr/bin/env bash

# Please note that this script only work with Github repositories.
# Prerequisites: git, github cli

GHCLI_BIN="ghcli"
REPO="cloudskiff/driftctl"
LATEST_TAG=$(git describe --abbrev=0) # Get the least created tag
DEFAULT_BRANCH=$(git symbolic-ref refs/remotes/origin/HEAD | sed 's@^refs/remotes/origin/@@')
BASE=$DEFAULT_BRANCH # Change this if you don't want to use the default branch as base

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

exit 0
