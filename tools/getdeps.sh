#!/bin/bash

set -e -x -u

scriptdir="$(realpath "$(dirname "$0")")"

# Path to the getdeps executable. If not specified, it will be
# built from local sources.
getdeps="${GETDEPS:-}"

if [ -z "${getdeps}" ]; then
  pushd "${scriptdir}"
  GO111MODULE=off go build -o getdeps/getdeps ./getdeps
  getdeps="${scriptdir}/getdeps/getdeps"
  popd
fi

exec "${getdeps}" "$@"
