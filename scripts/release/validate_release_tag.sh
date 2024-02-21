#!/usr/bin/env bash

set -o nounset
set -o errexit
set -E
set -o pipefail

CURRENT_RELEASE_TAG=$1

semver_pattern="^[0-9]+\.[0-9]+\.[0-9]+(-[a-z][a-z0-9]*)?$"

if ! [[ $1 =~ $semver_pattern ]]; then
    echo "Given tag \"$1\" does not match the expected semantic version pattern: \"${semver_pattern}\"."
    exit 1
fi

