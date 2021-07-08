#!/bin/bash -exu

scriptdir="$(realpath "$(dirname "$0")")"
OUT=${OUT:-"linuxboot_uroot_ttys0"}

artifacts="${ARTIFACTSDIR:-${scriptdir}/linuxboot-artifacts}"
kconfig="${KCONFIG:-${artifacts}/config.linuxboot.x86_64}"
patchdir="${PATCHDIR:-${scriptdir}/patches}"
initramfs=${INITRAMFS:-}
num_cores=$(nproc --all)
kernel_image="arch/x86/boot/bzImage"

pushd kernel

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
