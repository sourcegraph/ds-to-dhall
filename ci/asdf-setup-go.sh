#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")"/..
set -euxo pipefail

ASDF_GOLANG_DIRECTORY="$(asdf where golang)"

# https://docs.github.com/en/actions/reference/workflow-commands-for-github-actions#setting-an-environment-variable
echo "::set-env name=GOROOT::${ASDF_GOLANG_DIRECTORY}"
