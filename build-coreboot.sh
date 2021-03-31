#!/bin/bash
#
# This script builds coreboot images for the supported platforms. Currently
# deltalake, tiogapass, monolake, qemu-x86_64.
# This script expects a few environment variables, that are set in TARGETS:
# * `PLATFORM`: the name of the platform to build coreboot for. E.g. "qemu"
# * `KERNEL`, the path of the Linux kernel to use as coreboot primary payload.
#   There is no secondary payload, so make sure to embed your initramfs in the
#   kernel image.
#
# If you need to run a local build, just call this script by setting various
# environment variables:
#
# PLATFORM=your-platform-name \  # e.g. deltalake, tiogapass, monolake
# VER=your-version-string \  # e.g. 1.0. Only needed for deltalake, default is 0.0.0
# KERNEL=~/fbcode/buck-out/gen/osf/linuxboot/uroot-x86_64-outputs__srcs/kernel-source/arch/x86/boot/bzImage \
# ./build-coreboot.sh
#
# Optional variables:
# OUT: path to the output file. Default: coreboot-${PLATFORM}.rom
# PATCHDIR: path to the directory containing patch files. Each patch file name
#     has to end with ".patch" in order to be picked
#
# Remember to rebuild the kernel and ramfs in case you are making local changes:
#   buck build //osf/linuxboot:kernel

set -e -x -u

scriptdir="$(realpath "$(dirname "$0")")"
OUT=${OUT:-"coreboot-${PLATFORM}.rom"}
config="${scriptdir}/linuxboot-artifacts/coreboot-config-${PLATFORM}"
patchdir=${PATCHDIR:-${scriptdir}}
patches="${patchdir}/coreboot-${PLATFORM}-*.patch"
KERNEL=${KERNEL:-"${PWD}/kernel/linuxboot_uroot_ttys0"}
VER=${VER:-"0.0.0"}
VPD=${VPD:-"${scriptdir}/linuxboot-artifacts/vpd"}

pushd coreboot

check_gcc() {
  # This function verifies that gcc tool has version greater than or equal to 5.0.0.
  # gcc version 4.8.5 is known to not build.
  # gcc version 6.3.0 is known to build.
  # gcc version 7.0.0 is needed for recognizing -Wimplicit-fallthrough compile flag.
  currentver="$(gcc -dumpversion)"
  requiredver="7.0.0"
  # If gcc -dumpversion doesn't show subversion e.g. 7, then add two 0 subversion to it e.g. 7.0.0
  if [ "$(echo $currentver | grep -c '\.' )" -eq 0 ]; then
    currentver="$currentver.0.0"
  fi

  if [ "$(printf '%s\n' "$requiredver" "$currentver" | sort -V | head -n1)" = "$requiredver" ]
  then
    return 0
  else
    echo "GCC version is less than $requiredver"
    exit 1
  fi
}


apply_patches()
{
  for p in $patches
  do
    if [ -f "$p" ]
    then
      patch -p1 < "$p"
    fi
  done
}

setup_crossgcc()
{
  crossgcc_dir="util/crossgcc/"
  if [ -d "${crossgcc_dir}/xgcc" ]
  then
    echo "${crossgcc_dir}/xgcc exists, assuming crossgcc has already been built."
    return 0
  fi

  export EXTERNAL_DEPS_FILE="$(mktemp -t external_deps.XXXXX)"
  rm -f "${EXTERNAL_DEPS_FILE}"

  # Build the toolchain: this step takes a lot of time, and we should probably
  # use our internal gcc with patches, bells and whistles.
  if ! CPUS=$(nproc --ignore=1 --all) BUILD_LANGUAGES=c make "crossgcc-i386"; then
    set +x
    {
      echo
      echo "=== Toolchain build failed"
      if [ -f "${EXTERNAL_DEPS_FILE}" ]; then
        echo
        echo -n "One or more additional external dependencies are required for the build: "
        cat "${EXTERNAL_DEPS_FILE}"
        echo
        echo "Please re-run build with USE_FWDPROXY=1 to collect all the required files."
        rm -f "${EXTERNAL_DEPS_FILE}"
      fi
    } >&2
    exit 1
  fi
  if [ -f "${EXTERNAL_DEPS_FILE}" ]; then
    set +x
    {
      echo
      echo "The following additional external dependencies were required:"
      echo
      cat "${EXTERNAL_DEPS_FILE}"
      echo
      echo "Please check them into opsfiles_bin and update configs."
    } >&2
    rm -f "${EXTERNAL_DEPS_FILE}"
    # Fail the build so the message is not buried.
    exit 2
  fi
}

make_coreboot() {
  # Prepare .config and apply patches
  cp "${config}" ".config"

  # make sure that oldconfig won't go interactive. If `conf` succeeds, oldconfig
  # won't prompt for input.
  # ./build/util/kconfig/conf --listnewconfig src/Kconfig
  #timeout 10 make oldconfig
  sed -i "s|# CONFIG_ANY_TOOLCHAIN is not set|CONFIG_ANY_TOOLCHAIN=n|g" .config

  # The `kernel` variable is set via buck in TARGETS
  # shellcheck disable=SC2154
  sed -i "s|CONFIG_PAYLOAD_FILE=.*|CONFIG_PAYLOAD_FILE=\"${KERNEL}\"|g" .config

  # Try doing a parallel build first. If -jN fails, fall back to -j1 since it's
  # easier to read error output that way.
  #
  # FIXME: There has been some flakiness observed when doing many parallel make
  # jobs due to generated dependencies not showing up when they're supposed to.
  # We can investigate that to maximize parallelization, but for now just pick
  # a conservative number of jobs and fall back to -j1 if needed.
  make_jobs=4

  # UPDATED_SUBMODULES tells coreboot that the submodules are already
  # up-to-date. By default, coreboot will checkout the master branch for each
  # submodule. However we use branches e.g. for the blobs repo because it
  # contains binary blobs that we can't publish upstream.
  timeout 10 make olddefconfig || echo "Need to rebase .config manually"
  UPDATED_SUBMODULES=1 make -j${make_jobs} || UPDATED_SUBMODULES=1 make
  cp build/coreboot.rom "$OUT"
}

create_vpd_variables() {
  # the qemu image doesn't have a RO_VPD section in the flashmap yet
  if [ "$PLATFORM" = "qemu-x86_64" ]
  then
    echo "Skipping VPD creation for qemu-x86_64 because it has no VPD sections defined in its flashmap"
    return
  fi

  internal_versions="$(cat ../generated_versions.json)"
  $VPD -f "$OUT" -O -i RO_VPD -s internal_versions="$internal_versions"

  # Set overall firmware version if VER variable was defined as an input.
  # Effectively "buck build" does not set this.
  if [ "$VER" != "0.0.0" ]
  then
    $VPD -f "$OUT" -i RO_VPD -s firmware_version="$VER"
  fi

  # For DeltaLake, we need to set up several VPD variables.
  # These variables work for DeltaLake (but not other platfrom at the moment,
  # since they need support from firmware components, including FSP, coreboot, u-root.
  if [ "$PLATFORM" = "deltalake-evt" ] || [ "$PLATFORM" = "deltalake-dvt" ]
  then
    # Disable FSP log, set log level to 2:Warning.
    $VPD -f "$OUT" -i RO_VPD -s fsp_log_enable=0
    $VPD -f "$OUT" -i RO_VPD -s fsp_log_level=2
  fi
  # Initialize RW_VPD to empty region
  $VPD -f "$OUT" -O -i RW_VPD
}


check_gcc
apply_patches
setup_crossgcc
make_coreboot
create_vpd_variables

echo "Coreboot build done"
