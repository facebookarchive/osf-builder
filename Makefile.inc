# vim: syntax=make
#
# Copyright (c) Facebook, Inc. and its affiliates.
#
# This source code is licensed under the MIT license found in the
# LICENSE file in the root directory of this source tree.

# This file should be paramterized and included by platform Makefile.
# See `examples/qemu/Makefile` for one.

MAKEFLAGS += --warn-undefined-variables
.PHONY: always-build build clean clean-coreboot clean-initramfs clean-kernel coreboot-toolchain initramfs kernel wipe wipe-initramfs wipe-coreboot wipe-kernel
.DEFAULT_GOAL: build
.SUFFIXES:

ifndef PLATFORM
$(error Please define PLATFORM)
endif
ifndef OSF_BUILDER_DIR
$(error Please define OSF_BUILDER_DIR)
endif
ifndef CONFIGS_DIR
$(error Please define CONFIGS_DIR)
endif
ifndef PATCHES_DIR
$(error Please define PATCHES_DIR)
endif
# If defined, specifies URL overrides to be applied when fetching deps.
URL_OVERRIDES ?=
# Version of the firmware being built.
VERSION ?= 0.0.0

HASH_MODE ?= strict

BUILD_DIR ?= $(realpath $(CURDIR))/build
PLATFORM_BUILD_DIR ?= $(BUILD_DIR)/$(PLATFORM)
ROM_OUT ?= $(PLATFORM_BUILD_DIR)/osf-$(PLATFORM).rom

# Absolutize CONFIGS_DIR
CONFIGS_DIR := $(realpath $(CONFIGS_DIR))
CONFIG ?= $(CONFIGS_DIR)/config-$(PLATFORM).json
KERNEL_CONFIG ?= $(CONFIGS_DIR)/kernel.config
COREBOOT_CONFIG ?= $(CONFIGS_DIR)/coreboot.config-$(PLATFORM)
TOOLS_DIR := $(realpath $(OSF_BUILDER_DIR)/tools)
URL_OVERRIDES_ABS = $(realpath $(URL_OVERRIDES))

# We build initramfs every time to pick up possible changes.
# If you are not actively working on u-root, feel free to set this to 0.
ALWAYS_BUILD_INITRAMFS ?= 1
# Same for the kernel and coreboot.
ALWAYS_BUILD_KERNEL ?= 1
ALWAYS_BUILD_COREBOOT ?= 1
_WIPE_CONFIRMED ?= 0

ALL_CONFIGS := $(wildcard $(CONFIGS_DIR)/*.json)
DEFAULT_GETDEPS_TOOL ?= $(PLATFORM_BUILD_DIR)/getdeps
GETDEPS_TOOL ?= $(DEFAULT_GETDEPS_TOOL)
VPD_TOOL ?= $(TOOLS_DIR)/vpd
FLASHROM_TOOL ?= $(TOOLS_DIR)/flashrom
CBNT_PROV_TOOL ?= $(TOOLS_DIR)/cbnt-prov
FINAL_CONFIG_OUT := $(PLATFORM_BUILD_DIR)/final_config.json
NPROC ?= $(shell nproc --all --ignore=1)

INITRAMFS_BUILD_DIR := $(PLATFORM_BUILD_DIR)/initramfs
INITRAMFS_DEPS_FLAG := $(PLATFORM_BUILD_DIR)/.initramfs-deps
INITRAMFS_PATCH_FLAG := $(PLATFORM_BUILD_DIR)/.initramfs-patch
ifndef INITRAMFS_OUT
INITRAMFS_OUT := $(INITRAMFS_BUILD_DIR)/initramfs_linuxboot.amd64.cpio
endif
# By default, build initramfs but a prebuilt one can be provided via INITRAMFS_IN.
INITRAMFS_IN ?= $(INITRAMFS_OUT)

KERNEL_BUILD_DIR := $(PLATFORM_BUILD_DIR)/kernel
KERNEL_CONFIG_OUT := $(KERNEL_BUILD_DIR)/.config
KERNEL_DEPS_FLAG := $(PLATFORM_BUILD_DIR)/.kernel-deps
KERNEL_PATCH_FLAG := $(PLATFORM_BUILD_DIR)/.kernel-patch
KERNEL_OUT_DEFAULT := $(KERNEL_BUILD_DIR)/arch/x86/boot/bzImage
ifndef KERNEL_OUT
KERNEL_OUT := $(KERNEL_OUT_DEFAULT)
endif
# Prebuilt kernel can be provided via KERNEL_IN.
KERNEL_IN ?= $(KERNEL_OUT)

COREBOOT_BUILD_DIR := $(PLATFORM_BUILD_DIR)/coreboot
COREBOOT_CONFIG_OUT := $(COREBOOT_BUILD_DIR)/.config
COREBOOT_TOOLCHAIN_OUT := $(COREBOOT_BUILD_DIR)/util/crossgcc/xgcc
COREBOOT_OUT := $(COREBOOT_BUILD_DIR)/build/coreboot.rom
COREBOOT_DEPS_FLAG := $(PLATFORM_BUILD_DIR)/.coreboot-deps
COREBOOT_PATCH_FLAG := $(PLATFORM_BUILD_DIR)/.coreboot-patch
COREBOOT_TOOLCHAIN_FLAG := $(PLATFORM_BUILD_DIR)/.coreboot-toolchain
COREBOOT_TOOLCHAIN_CACHE_ENABLE ?= 1
# coreboot Intel blobs
ifdef COREBOOT_BLOBS_DIR
COREBOOT_BLOBS_DIR := $(realpath $(COREBOOT_BLOBS_DIR))
COREBOOT_FSP_T := $(COREBOOT_BLOBS_DIR)/Server_T.fd
COREBOOT_FSP_M := $(COREBOOT_BLOBS_DIR)/Server_M.fd
COREBOOT_FSP_S := $(COREBOOT_BLOBS_DIR)/Server_S.fd
COREBOOT_UCODE := $(COREBOOT_BLOBS_DIR)/microcode.mcb
COREBOOT_IFD := $(COREBOOT_BLOBS_DIR)/flashregion_0_flashdescriptor.bin
COREBOOT_IGN := $(COREBOOT_BLOBS_DIR)/flashregion_2_intel_me.bin
endif

UROOT_BASE_CMDS ?= \
	boot/fbnetboot \
	boot/localboot \
	boot/systemboot \
	core/cat \
	core/chmod \
	core/chroot \
	core/cmp \
	core/cp \
	core/date \
	core/dhclient \
	core/dd \
	core/df \
	core/dirname \
	core/dmesg \
	core/echo \
	core/elvish \
	core/find \
	core/free \
	core/grep \
	core/hostname \
	core/id \
	core/init \
	core/insmod \
	core/ip \
	core/kexec \
	core/kill \
	core/ln \
	core/ls \
	core/lsmod \
	core/mkdir \
	core/mknod \
	core/mount \
	core/mv \
	core/ntpdate \
	core/ping \
	core/ps \
	core/rm \
	core/rmmod \
	core/sleep \
	core/shutdown \
	core/sync \
	core/tail \
	core/tee \
	core/umount \
	core/uname \
	core/wget \
	exp/cbmem \
	exp/dmidecode \
	exp/modprobe \
	exp/ipmidump

UROOT_ADDITIONAL_CMDS ?=

UROOT_ADDITIONAL_GOPATH ?=

BASE_FILES ?= $(FLASHROM_TOOL):bin/flashrom $(VPD_TOOL):bin/vpd

UROOT_ADDITIONAL_FILES ?=

UINIT_CMD ?= systemboot

# Shorthand for building the firmware.
build: $(ROM_OUT)

# Main output product - firmware with VPD variables.
$(ROM_OUT): $(COREBOOT_OUT)
	$(VPD_TOOL) -f "$<" -O -i RO_VPD -s internal_versions="`cat $(FINAL_CONFIG_OUT)`"
	[ -z "$(VERSION)" ] || $(VPD_TOOL) -f "$<" -i RO_VPD -s firmware_version=$(VERSION)
	$(VPD_TOOL) -f "$<" -O -i RW_VPD || true # Format RW_VPD region, if present.
	cp "$<" "$@"
	@echo '***' >&2
	@echo '*** Build done, $@' >&2
	@echo '***' >&2

# ========================
# === coreboot section ===
# ========================

# coreboot build - produces a firmware image, depends on payload (kernel + initramfs), config and toolchain.
$(COREBOOT_OUT): $(KERNEL_IN) $(COREBOOT_CONFIG_OUT) $(COREBOOT_PATCH_FLAG) $(COREBOOT_TOOLCHAIN_FLAG)
	@# Parallel builds are not always reliable, hence -j1.
	UPDATED_SUBMODULES=1 MAKEFLAGS= $(MAKE) -C $(COREBOOT_BUILD_DIR) -j1

ifeq "$(ALWAYS_BUILD_COREBOOT)" "1"
$(COREBOOT_OUT): always-build
endif

# Applies coreboot patches for this platform, if any.
$(COREBOOT_PATCH_FLAG): $(realpath $(wildcard $(PATCHES_DIR)/coreboot-$(PLATFORM)-*))
	$(call patch,$(COREBOOT_BUILD_DIR),$^)
	touch $@

# Toolchain version identifier - same as in coreboot/util/crossgcc/buildgcc.
# Produces something like "2021-04-06_7014f825", uniquely identifes the toolchain and is used to cache it.
CROSSGCC_VERSION = $(shell git -C $(COREBOOT_BUILD_DIR)/util/crossgcc log -n 1 --pretty=%cd --date=short .)_$(shell git -C $(COREBOOT_BUILD_DIR)/util/crossgcc log -n 1 --pretty=%h .)
COREBOOT_TOOLCHAIN_CACHE = $(BUILD_DIR)/coreboot-toolchain-$(CROSSGCC_VERSION).tar

# Flag signaling that coreboot toolchain build is done.
# Depends on IASL binary which is built last during the toolchain build, meaning it's complete (in case build was interrupted).
# Toolchain cache archive is created here, unless already exists.
$(COREBOOT_TOOLCHAIN_FLAG): $(COREBOOT_TOOLCHAIN_OUT)/bin/iasl
ifeq "$(COREBOOT_TOOLCHAIN_CACHE_ENABLE)" "1"
	if ! [ -f "$(COREBOOT_TOOLCHAIN_CACHE)" ]; then \
	  tar cf $(COREBOOT_TOOLCHAIN_CACHE).tmp -C $(dir $(COREBOOT_TOOLCHAIN_OUT)) $(notdir $(COREBOOT_TOOLCHAIN_OUT)) && \
	  mv $(COREBOOT_TOOLCHAIN_CACHE).tmp $(COREBOOT_TOOLCHAIN_CACHE); \
	fi
endif
	touch $@

# Provides the crossgcc toolchain. Uses cache if enabled and available, otherwise build from sources.
$(COREBOOT_TOOLCHAIN_OUT)/bin/iasl: $(COREBOOT_DEPS_FLAG) $(COREBOOT_PATCH_FLAG)
	if [ -f $(COREBOOT_TOOLCHAIN_CACHE) ]; then \
	  echo Using cached toolchain from $(COREBOOT_TOOLCHAIN_CACHE) >&2; \
	  tar xf $(COREBOOT_TOOLCHAIN_CACHE) -C $(dir $(COREBOOT_TOOLCHAIN_OUT)); \
	  touch $@; \
	else \
	  COREBOOT_BUILD_DIR=$(COREBOOT_BUILD_DIR) CPUS=$(NPROC) MAKE=$(MAKE) TOOLS_DIR=$(TOOLS_DIR) \
	    $(TOOLS_DIR)/build_toolchain.sh; \
	fi

# Creates coreboot config file using the user-provided one and feeding it to make olddefconfig.
$(COREBOOT_CONFIG_OUT): $(COREBOOT_CONFIG) $(COREBOOT_DEPS_FLAG) $(COREBOOT_TOOLCHAIN_FLAG)
	cp $(COREBOOT_CONFIG) $(COREBOOT_CONFIG_OUT)
	if [ -d "$(COREBOOT_BLOBS_DIR)" ]; then \
	  echo Using provided coreboot Intel blobs FSP, ME and microcode from $(COREBOOT_BLOBS_DIR) >&2; \
	  sed -i "s|CONFIG_CPU_UCODE_BINARIES=.*|CONFIG_CPU_UCODE_BINARIES=\"$(COREBOOT_UCODE)\"|" $@; \
	  sed -i "s|CONFIG_FSP_T_FILE=.*|CONFIG_FSP_T_FILE=\"$(COREBOOT_FSP_T)\"|" $@; \
	  sed -i "s|CONFIG_FSP_M_FILE=.*|CONFIG_FSP_M_FILE=\"$(COREBOOT_FSP_M)\"|" $@; \
	  sed -i "s|CONFIG_FSP_S_FILE=.*|CONFIG_FSP_S_FILE=\"$(COREBOOT_FSP_S)\"|" $@; \
	  sed -i "s|CONFIG_IFD_BIN_PATH=.*|CONFIG_IFD_BIN_PATH=\"$(COREBOOT_IFD)\"|" $@; \
	  sed -i "s|CONFIG_ME_BIN_PATH=.*|CONFIG_ME_BIN_PATH=\"$(COREBOOT_IGN)\"|" $@; \
	fi
	sed -i "s|CONFIG_PAYLOAD_FILE=.*|CONFIG_PAYLOAD_FILE=\"$(KERNEL_IN)\"|" $@
	! grep -q CONFIG_INTEL_CBNT_PROV_EXTERNAL_BIN=y $@ || \
	  sed -i "s|CONFIG_INTEL_CBNT_PROV_EXTERNAL_BIN_PATH=.*|CONFIG_INTEL_CBNT_PROV_EXTERNAL_BIN_PATH=\"$(CBNT_PROV_TOOL)\"|" $@
	MAKEFLAGS= $(MAKE) -C $(COREBOOT_BUILD_DIR) olddefconfig

clean-coreboot:
	rm -f $(COREBOOT_OUT)
	[ -d $(COREBOOT_BUILD_DIR) ] && MAKEFLAGS= $(MAKE) -C $(COREBOOT_BUILD_DIR) clean

wipe-coreboot:
	rm -rf $(COREBOOT_BUILD_DIR) $(COREBOOT_DEPS_FLAG) $(COREBOOT_PATCH_FLAG) $(COREBOOT_TOOLCHAIN_FLAG)

# ======================
# === Kernel section ===
# ======================

# Top-level initramfs target, for use by humans.
kernel: $(KERNEL_OUT)

# Kernel build target. Requires initrmafs image, config and patches.
ifneq "$(INITRAMFS_IN)" ""
$(KERNEL_OUT): $(INITRAMFS_IN) $(KERNEL_CONFIG_OUT) $(KERNEL_PATCH_FLAG)
else
$(KERNEL_OUT): $(KERNEL_CONFIG_OUT) $(KERNEL_PATCH_FLAG)
endif
	@# If lzma(1) is not available, kernel build will fall back to gzip and we don't want that.
	@if ! lzma -h > /dev/null 2>/dev/null; then \
	  echo ' *** Please install the lzma CLI utility.' >&2; \
	  echo ' *** In RedHat distros it`s provided by xz-lzma-compat, in Debian/Ubuntu it`s provided by xz-utils.' >&2; \
	  exit 1; \
	fi
	MAKEFLAGS= $(MAKE) -C $(KERNEL_BUILD_DIR) -j$(NPROC) KCFLAGS=-pipe
	[ "$@" = "$(KERNEL_OUT_DEFAULT)" ] || cp -v "$(KERNEL_OUT_DEFAULT)" "$@"

ifeq "$(ALWAYS_BUILD_KERNEL)" "1"
$(KERNEL_OUT): always-build
endif

$(KERNEL_CONFIG_OUT): $(KERNEL_CONFIG) $(KERNEL_DEPS_FLAG)
	cp $(KERNEL_CONFIG) $@
	sed -i "s|CONFIG_INITRAMFS_SOURCE=.*|CONFIG_INITRAMFS_SOURCE=\"$(INITRAMFS_IN)\"|" $@
	MAKEFLAGS= $(MAKE) -C $(KERNEL_BUILD_DIR) olddefconfig

$(KERNEL_PATCH_FLAG): $(realpath $(wildcard $(PATCHES_DIR)/kernel-*))
	$(call patch,$(KERNEL_BUILD_DIR),$^)
	touch $@

clean-kernel:
	rm -f $(KERNEL_OUT)
	[ -d $(KERNEL_BUILD_DIR) ] && MAKEFLAGS= $(MAKE) -C $(KERNEL_BUILD_DIR) clean

wipe-kernel:
	rm -rf $(KERNEL_BUILD_DIR) $(KERNEL_DEPS_FLAG) $(KERNEL_PATCH_FLAG)

# =========================
# === Initramfs section ===
# =========================

# Top-level initramfs target, for use by humans.
initramfs: $(INITRAMFS_OUT)

# Initramfs build target.
$(INITRAMFS_OUT): $(INITRAMFS_DEPS_FLAG) $(INITRAMFS_PATCH_FLAG)
	GO111MODULE=off \
	GOROOT=$(INITRAMFS_BUILD_DIR)/go \
	GOCACHE=$(INITRAMFS_BUILD_DIR)/go/.cache \
	GOPATH=$(INITRAMFS_BUILD_DIR)/gopath:$(UROOT_ADDITIONAL_GOPATH) \
	  $(INITRAMFS_BUILD_DIR)/go/bin/go run github.com/u-root/u-root \
	    -build=bb -o $@ -uinitcmd=$(UINIT_CMD) \
	    $(addprefix -files=,$(BASE_FILES) $(UROOT_ADDITIONAL_FILES)) \
	    $(addprefix github.com/u-root/u-root/cmds/,$(UROOT_BASE_CMDS)) $(UROOT_ADDITIONAL_CMDS) \
	  || { rm -f $@; exit 1; }

ifeq "$(ALWAYS_BUILD_INITRAMFS)" "1"
$(INITRAMFS_OUT): always-build
endif

$(INITRAMFS_PATCH_FLAG): $(realpath $(wildcard $(PATCHES_DIR)/initramfs-*))
	$(call patch,$(INITRAMFS_BUILD_DIR),$^)
	touch $@

clean-initramfs:
	rm -f $(INITRAMFS_OUT)

wipe-initramfs:
	rm -rf $(INITRAMFS_BUILD_DIR) $(INITRAMFS_DEPS_FLAG) $(INITRAMFS_PATCH_FLAG)

# ====================
# === Misc section ===
# ====================

# If any of the configs is not there, give the user a clear message (probably PLATFORM is incorrect).
$(CONFIG) $(COREBOOT_CONFIG) $(KERNEL_CONFIG):
	@echo '***' >&2
	@echo '*** Missing config - $@' >&2
	@echo '***' >&2
	@false

# Rebuild the getdeps tool if necessary.
$(DEFAULT_GETDEPS_TOOL): $(wildcard $(OSF_BUILDER_DIR)/getdeps/*.go)
	mkdir -p $(PLATFORM_BUILD_DIR)
	GO111MODULE=off GOBIN=$(dir $@) go get $(OSF_BUILDER_DIR)/getdeps

# Run the getdeps tool for a component (initramfs, kernel, coreboot) and create a flag file.
# The flag file is used to avoid re-running unless JSON configs have changed.
$(PLATFORM_BUILD_DIR)/.%-deps: $(CONFIG) $(GETDEPS_TOOL) $(ALL_CONFIGS)
	@if [ -f "$@" -a $(_WIPE_CONFIRMED) -eq 0 ]; then \
	  echo "WARNING: A full wipe is being triggered, it will delete all the downloaded code under $(PLATFORM_BUILD_DIR), are you sure to continue? (y/n)"; \
	  read ans; \
	  if [ "$$ans" = "y" ]; then \
	    echo continue; \
	  else \
	    echo exit 1; exit 1; \
	  fi \
	fi
	$(eval _WIPE_CONFIRMED=1)
	mkdir -p $(PLATFORM_BUILD_DIR)
	cd $(PLATFORM_BUILD_DIR) && $(GETDEPS_TOOL) --components $* -c $(CONFIG) -H $(HASH_MODE) --url-overrides=$(URL_OVERRIDES_ABS) -o $(FINAL_CONFIG_OUT)
	touch $@

define patch  # dir,patches
	[ -z "$2" ] || { cd $1 && for p in $2; do patch -p 1 -N -b --verbose -i $$p; done }
endef

clean: clean-initramfs clean-kernel clean-coreboot
	rm -f $(OUT)

wipe:
	rm -rf $(PLATFORM_BUILD_DIR)

# Dummy target to depend on when forcing rebuild.
always-build:
	@true
