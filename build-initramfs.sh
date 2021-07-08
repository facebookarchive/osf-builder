#!/bin/bash -exu
# This is a script to build an u-root initramfs. It expects the go compiler to
# be placed inside initramfs/go and the go libraries inside initramfs/go/gopath

scriptdir="$(realpath "$(dirname "$0")")"
OUT=${OUT:-"initramfs_linuxboot.amd64.cpio"}

pushd initramfs
export PATH="${PWD}/go/bin:${PATH}"

# Set up GOROOT and GOPATH.
export GOROOT="${PWD}/go"
GOPATH="${PWD}/gopath"
# Go modules disabled since we build from pinned commits from a custom GOPATH
export GO111MODULE=off
# additional colon-separated GOPATH entries for additional components
ADDITIONAL_GOPATH=${ADDITIONAL_GOPATH:-}
if [ -n "${ADDITIONAL_GOPATH}" ]
then
    echo "Using additional GOPATH components: ${ADDITIONAL_GOPATH}"
    GOPATH="${GOPATH}:${ADDITIONAL_GOPATH}"
fi
export GOPATH

# Apply patches.
patchdir="${PATCHDIR:-${scriptdir}/patches}"
for p in "${patchdir}"/initramfs-*.patch; do
  p=$(realpath $p)
  echo "Applying patch: $p"
  patch -d gopath/src/github.com/u-root/u-root -p 1 -b < "$p"
done

ADDITIONAL_CMDS=${ADDITIONAL_CMDS:-}
UINITCMD=${UINITCMD-systemboot}

base_cmds=(
  "boot/localboot"
  "boot/fbnetboot"
  "boot/systemboot"
  "core/cat"
  "core/chmod"
  "core/chroot"
  "core/cmp"
  "core/cp"
  "core/date"
  "core/dhclient"
  "core/dd"
  "core/df"
  "core/dirname"
  "core/dmesg"
  "core/echo"
  "core/elvish"
  "core/find"
  "core/free"
  "core/grep"
  "core/hostname"
  "core/id"
  "core/init"
  "core/insmod"
  "core/ip"
  "core/kexec"
  "core/kill"
  "core/ln"
  "core/ls"
  "core/lsmod"
  "core/mkdir"
  "core/mknod"
  "core/mount"
  "core/mv"
  "core/ntpdate"
  "core/ping"
  "core/ps"
  "core/rm"
  "core/rmmod"
  "core/sleep"
  "core/shutdown"
  "core/sync"
  "core/tail"
  "core/tee"
  "core/umount"
  "core/uname"
  "core/wget"
  "exp/cbmem"
  "exp/dmidecode"
  "exp/modprobe"
  "exp/ipmidump"
)

flags=()

if [ -n "${UINITCMD}" ]
then
    echo "Init command: ${UINITCMD}"
    flags=("${flags[@]}" "-uinitcmd=${UINITCMD}")
fi

flags=("${flags[@]}"
  "-files" "$(readlink -f "${scriptdir}"/resources/flashrom):bin/flashrom"
  "-files" "$(readlink -f "${scriptdir}"/resources/vpd):bin/vpd"
)

for cmd in "${base_cmds[@]}"
do
  uroot_cmds=("${uroot_cmds[@]}" "github.com/u-root/u-root/cmds/${cmd}")
done

for cmd in $ADDITIONAL_CMDS
do
  uroot_cmds=("${uroot_cmds[@]}" "${cmd}")
done

go build github.com/u-root/u-root

echo "Generating initramfs to ${OUT}"
./u-root -build=bb -o "${OUT}" "${flags[@]}" "${uroot_cmds[@]}"

echo "Initramfs build done"
