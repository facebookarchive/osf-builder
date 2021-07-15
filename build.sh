#!/bin/bash
#
# This is the build script.
#
# During the build process, it removes and rebuilds
# these directories: initramfs, kernel, coreboot.
#
# Make sure the gcc version is higher than 4.8
#
# The following environment variables should be set:
#
# PLATFORM=ac|qemu-x86_64

set -e -x -u

scriptdir="$(realpath "$(dirname "$0")")"

export CONFIGDIR=${CONFIGDIR:-${scriptdir}/configs}
export HASH_MODE=${HASH_MODE:-strict}
export PLATFORM=${PLATFORM:-ac}

# Build initramfs based on go/u-root
"${scriptdir}/build-initramfs.sh"

# Build Linux kernel and then coreboot payload
"${scriptdir}/build-kernel.sh"

# Build coreboot and then Linuxboot FW image
"${scriptdir}/build-coreboot.sh"
