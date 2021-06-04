#!/usr/bin/env bash

COMMIT_MESSAGE="$(git log --format=oneline -n 1 $CIRCLE_SHA1)"

if [[ $COMMIT_MESSAGE == *"[RUN ACC]"* ]] || [[ $FORCE_ACC_TEST == "true" ]]; then
  # If user is not member of core team, context is not applie so acc test won't run anyway
  echo "Running acceptance tests"
else
  echo "Not running acceptance tests"
  circleci-agent step halt
fi
