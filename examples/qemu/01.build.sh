#!/bin/bash
set -exu

# build OSF for QEMU. The resulting image will be located at
# coreboot/build/coreboot.rom .
PLATFORM=qemu-x86_64 CONFIGDIR=${PWD}/configs PATCHDIR=${PWD}/patches RESOURCESDIR=${PWD}/resources ARTIFACTSDIR=${PWD}/artifacts ../../build.sh
