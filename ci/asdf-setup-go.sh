#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/..
set -euxo pipefail

ASDF_GOLANG_DIRECTORY="$(asdf where golang)/go"

# https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#environment-files
echo "GOROOT=${ASDF_GOLANG_DIRECTORY}" >>"$GITHUB_ENV"
