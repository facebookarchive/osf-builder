#!/bin/bash
set -exu

# remove generate_versions.json which was created by getdeps
rm -f generated_versions.json

# remove osf-builder/kernel/initramfs/coreboot directories
rm -rf osf-builder
rm -rf kernel
rm -rf initramfs
rm -rf coreboot
