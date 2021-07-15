#!/bin/bash -exu

scriptdir="$(realpath "$(dirname "$0")")"
OUT=${OUT:-"${scriptdir}/kernel/linuxboot_uroot_ttys0"}
CONFIGDIR=${CONFIGDIR:-${scriptdir}/configs}

kconfig="${KCONFIG:-${CONFIGDIR}/kernel-linuxboot.config}"
patchdir="${PATCHDIR:-${scriptdir}/patches}"
initramfs=${INITRAMFS:-${scriptdir}/initramfs/initramfs_linuxboot.amd64.cpio}
num_cores=$(nproc --all)
kernel_image="arch/x86/boot/bzImage"

NODEPS=${NODEPS:-0}
HASH_MODE=${HASH_MODE:-strict}
if [ "${NODEPS}" != "1" ]; then
  "${scriptdir}/tools/getdeps.sh" --components kernel -c "${CONFIGDIR}/config-${PLATFORM}.json" -H "${HASH_MODE}"
fi

pushd kernel

# Snapshots come with a linux-x.y.z directory, git clones don't have it, we account for both.
[ -e Makefile ] || pushd linux-*.*.*

echo "Applying patches from ${patchdir}"
for p in "${patchdir}"/kernel-*; do
  # using cat so the patch names show up in the logs because of set -x
  # shellcheck disable=SC2002
  cat "${p}" | patch -p1
done

cp "${kconfig}" "${PWD}/.config"
if [ -n "${initramfs}" ]
then
    sed -i "s|^CONFIG_INITRAMFS_SOURCE=.*|CONFIG_INITRAMFS_SOURCE=\"${initramfs}\"|" .config
fi

# If lzma(1) is not available, kernel build will fall back to gzip and we don't want that.
if ! lzma -h > /dev/null 2>/dev/null; then
  echo 'Please install the lzma CLI utility (in RedHat distros it`s provided by xz-lzma-compat)'
  exit 1
fi

make clean
timeout 60 make olddefconfig
make -j$((num_cores+1)) KCFLAGS="-pipe"

cp "${kernel_image}" "${OUT}"

echo "Kernel build done"
