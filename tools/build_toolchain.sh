#!/bin/bash
#
# Copyright (c) Facebook, Inc. and its affiliates.
#
# This source code is licensed under the MIT license found in the
# LICENSE file in the root directory of this source tree.

set -e -u

ALLOW_EXTERNAL_DEPS=${ALLOW_EXTERNAL_DEPS:-0}

EXTERNAL_DEPS_FILE="$(mktemp -t external_deps.XXXXX)"
function cleanup() {
  rm -f "${EXTERNAL_DEPS_FILE}"
}
trap cleanup EXIT
cleanup

# This will be used by the wget script to record external dependencies while fetching.
export EXTERNAL_DEPS_FILE

if ! MAKEFLAGS="" PATH="${TOOLS_DIR}:${PATH}" "${MAKE}" -C "${COREBOOT_BUILD_DIR}" crossgcc-i386 BUILD_LANGUAGES=c; then
  {
    echo
    echo "=== Toolchain build failed"
    if [ -f "${EXTERNAL_DEPS_FILE}" ]
    then
      echo
      echo -n "One or more additional external dependencies are required for the build: "
      cat "${EXTERNAL_DEPS_FILE}"
      echo
      echo "Please re-run build with USE_FWDPROXY=1 to collect all the required files."
    fi
  } >&2
  exit 1
fi

if [ -f "${EXTERNAL_DEPS_FILE}" ]
then
  {
    echo
    echo "The following additional external dependencies were required:"
    echo
    cat "${EXTERNAL_DEPS_FILE}"
    echo
    echo "Please update configs."
  } >&2
  if [ "${ALLOW_EXTERNAL_DEPS}" != "1" ]; then
    # Fail the build so the message is not buried.
    exit 2
  fi
fi
