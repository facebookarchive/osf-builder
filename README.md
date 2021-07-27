# OSF Builder

This osf-builder repo is an open source repo to faciliate the development
community to collaborate on Open System Firmware (OSF).

## Build pre-requisites

 * GNU make
 * Go 1.x

## Build process

OSF boot starts with [coreboot](https://coreboot.org/) first, then [Linux kernel](https://kernel.org/) which then executes init, which in this case is provided by [u-root](https://github.com/u-root/u-root).

OSF build also consists of these three stages, executed in the reverse order.

Entire build requires `PLATFORM` to be defined, this specifies the platform for which build is being run.

 * `getdeps` is a tool used to fetch dependencies. It can clone Git repos, fetch files, etc.
   * It is configured by a JSON file that must be specified in `CONFIG`.
   * `CONFIG` consists of three top-level sections: `initramfs`, `kernel` and `coreboot` that specify what to fetch for each of the stages.
 * Initramfs image is built first, by building u-root with certain set of commands.
   * `initramfs` section of the `CONFIG` is executed by `getdeps` to fetch the u-root sources and the Go toolchain.
   * `PATCHES_DIR/initramfs-PLATFORM-*` patches are applied.
   * Default set of commands can be found in Makefile.inc `UROOT_BASE_CMDS`, it can be augmented with `UROOT_ADDITIONAL_CMDS` or replaced entirely.
   * Additional commands can come from u-root itself or from external packages, in which case `UROOT_ADDITIONAL_GOPATH` may be required.
   * Initramfs can embed binary utilities, files can be added through `UROOT_ADDITIONAL_FILES` as `local_path:initramfs_path` pairs.
 * Kernel is built next
   * `kernel` section of the `CONFIG` is executed by `getdeps` to fetch the kernel source.
   * `PATCHES_DIR/kernel-PLATFORM-*` patches are applied.
   * `KERNEL_CONFIG` is used as `.config`.
 * Coreboot is built last
   * `coreboot` section of the `CONFIG` is executed by `getdeps` to fetch the source and toolchain dependencies.
   * `PATCHES_DIR/coreboot-PLATFORM-*` patches are applied.
   * `COREBOOT_CONFIG` is used as `.config`.
   * Resulting flahs image is written to `osf-PLATFORM.rom` in the current directory.

## How to build the sample project

* Clone the repo.
* cd examples/qemu
* Run `make`
* Once the build is completed, run `make run`, it will start a VM with the OSF BIOS image.

## Development tricks

 * To speed up builds, when not actively working on initramfs or the kernel, pass `ALWAYS_BUILD_INITRAMFS=0` and `ALWAYS_BUILD_KERNEL=0` respectively.
   * `make ALWAYS_BUILD_INITRAMFS=0 ALWAYS_BUILD_KERNEL=0` - for hacking on coreboot only.
 * `make clean` will clean all the components without wiping the work done by `getdeps`.
   * `make clean-coreboot` and `make clean-kernel` will clean just the coreboot and kernel components.
 * `make wipe` will wipe everything, including downloaded deps.
   * `make wipe-coreboot` and `make wipe-kernel` will clean just the coreboot and kernel components.
   * Note that toolchain cache survives wipe and will be used in the next build.

## License

OSF Builder is MIT licensed, as found in the LICENSE file.
