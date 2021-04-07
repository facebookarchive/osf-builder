#!/bin/bash -exu
#
# This is the top level script to build without kernel and u-root
# source code being in FBCODE.
# It needs to be run under Linuxboot root directory.
# During the build process, it removes and rebuilds
# these directories: initramfs, kernel-source, coreboot.
#
# Make sure the gcc version is higher than 4.8
#
# The following environment variables should be set:
#
# PLATFORM=deltalake (or tiogapass, monolake)
# CONFIGDIR=/path/to/dir: optional, path to where config-$PLATFORM.json is
#     located. Default: ${scriptdir}/configs/
# VER=<overall firmware version>: optional, effective for deltalake
# KM_PRIV_KEY_PATH=<path to KM private key>: optional, effective for deltalake-dvt
# ODM_PRIV_KEY_PATH=<path to ODM private key>: optional, effective for deltalake-dvt
#     KM_PRIV_KEY_PATH and ODM_PRIV_KEY_PAH need to be provided together, if neither KM_PRIV_KEY_PATH nor
# ODM_PRIV_KEY_PATH is provided for deltalake-dvt, it would use the default test keys
#     under cbnt to sign coreboot image.

if [ -z "$PLATFORM" ]
then
  echo "PLATFORM environment variable is undefined."
  exit 1
fi

scriptdir="$(realpath "$(dirname "$0")")"

CONFIGDIR=${CONFIGDIR:-${scriptdir}/configs/}
HASH="${HASH:-strict}"
KM_PRIV_KEY_PATH=${KM_PRIV_KEY_PATH:-""}
ODM_PRIV_KEY_PATH=${ODM_PRIV_KEY_PATH:-""}

# Path to the getdeps executable. If not specified, it will be
# built from local sources.
GETDEPS="${GETDEPS:-}"

if [ "${GETDEPS}" = "" ]
then
    pushd "${scriptdir}"
    go build -o cmd/getdeps/getdeps ./cmd/getdeps/
    GETDEPS="${scriptdir}/cmd/getdeps/getdeps"
    popd
fi

# Build initramfs based on go/u-root
"${GETDEPS}" --components initramfs -c "${CONFIGDIR}/config-${PLATFORM}.json"
"${scriptdir}/build-initramfs.sh"

# Build Linux kernel and then coreboot payload
"${GETDEPS}" --components kernel -c "${CONFIGDIR}/config-${PLATFORM}.json"
"${scriptdir}/build-kernel.sh"

# Build coreboot and then Linuxboot FW image
"${GETDEPS}" --components coreboot -c "${CONFIGDIR}/config-${PLATFORM}.json" -H "$HASH"
"${scriptdir}/build-coreboot.sh"
if [ "$PLATFORM" = "deltalake-dvt" ]
then
  if [ -z "$KM_PRIV_KEY_PATH" ] && [ -z "$ODM_PRIV_KEY_PATH" ]
  then
    KM_PRIV_KEY_PATH="cbnt/km_test_priv_key.pem"
    ODM_PRIV_KEY_PATH="cbnt/bpm_test_priv_key.pem"
    echo "Use default test keys $KM_PRIV_KEY_PATH and $ODM_PRIV_KEY_PATH"
  elif [ -z "$KM_PRIV_KEY_PATH" ] || [ -z "$ODM_PRIV_KEY_PATH" ]
  then
    echo "Both KM_PRIV_KEY_PATH and ODM_PRIV_KEY_PATH environment variables must be set."
    exit 1
  else
    echo "Use provided keys $KM_PRIV_KEY_PATH and $ODM_PRIV_KEY_PATH"
  fi
  PLATFORM=$PLATFORM KM_PRIV_KEY_PATH=$KM_PRIV_KEY_PATH ODM_PRIV_KEY_PATH=$ODM_PRIV_KEY_PATH ./build-cbnt.sh
fi
