#!/bin/bash
set -exu

# build OSF for QEmu. The resulting image will be located at
# coreboot/coreboot.rom .
PLATFORM=qemu-x86_64 CONFIGDIR=../../configs ../../build.sh
