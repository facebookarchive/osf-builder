#!/bin/bash
set -exu

# remove generate_versions.json which was created by getdeps
rm generated_versions.json

# remove kernel/initramfs/coreboot directories
rm -rf kernel
rm -rf initramfs
rm -rf coreboot
