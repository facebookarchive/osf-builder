#
# Copyright (c) Facebook, Inc. and its affiliates.
#
# This source code is licensed under the MIT license found in the
# LICENSE file in the root directory of this source tree.

PLATFORM := qemu-x86_64

OSF_BUILDER_DIR := ../..

# CONFIG, KERNEL_CONFIG and COREBOOT_CONFIG are derived from CONFIGS_DIR and PLATFORM.
CONFIGS_DIR := ./configs
PATCHES_DIR := ./patches

include $(OSF_BUILDER_DIR)/Makefile.inc

run:
	sudo qemu-system-x86_64 \
    `# the machine type specified in the coreboot mainboard configuration` \
    -M q35 \
    `# use KVM to avail of hardware virtualization extensions` \
    -enable-kvm \
    `# the coreboot ROM to run as system firmware` \
    -bios $(OUT) \
    `# the amount of RAM in MB` \
    -m 1024 \
    `# RNG to avoid DHCP lockups when waiting for entropy` \
    -object rng-random,filename=/dev/urandom,id=rng0 \
    `# redirect all the output to the console` \
    -nographic
