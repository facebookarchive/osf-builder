osf-builder

This osf-builder repo is an open source repo to faciliate the development
community to collaborate on Open System Firmware (OSF).

The OSF boot process has following phases:
1. Coreboot. This component initize SoC and platform component (such as memory)
to the extent that kernel on flash can be loaded. In future, UEFI support may be
added.
2. Linux Kernel. This component includes device drivers, file system drivers
and networking drivers to support local boot and network boot. The kernel binary
is built with Linux kernel code using Linuxboot specific Kconfig, which is needed
to reduce kernel size.
3. Initramfs. This component give system control to the target OS either locally
or remotely. The initramfs is u-root, which is similar to busybox. It is built
with go language.

In the build process, kernel is built first, followed by initramfs, and coreboot.
In the end, flashable OSF host firmware image is built.

This repo has 3 components:
a. A code syncer. It is in cmd directory. It is written in go. It builds into
getdpes binary, which is able to sync down code base based on JSON format
configuration. The code base may be obtained from a certain commit of a certain
branch of a certain repo, or from a local compressed tar ball, or from a URL.
b. Build scripts. The build scripts get source code, build initramfs, kernel
and then coreboot.
c. A sample project. The examples/qemu directory contains a sample project.
A project may have following components:
** scripts to build and to clean. The build script may set up appropriate
environment variables and call the osf-builder build script.
** Configs: The code base configuration in JSON format. The JSON config files
can be nested.
** Artifacts: Linuxboot kernel Kconfig, coreboot config.
** Patches: Patch files to be applied to kernel, u-root and coreboot.
** Resources: Local tar balls and static binaries to be included in initramfs.

Following are build server dependencies:
* Compilers and build tools - These include jq, vpd tools, go compilers.
* gcc version needs to be greater than or equal to 5.0.0. 

## How to build the sample project
* Clone the repo.
* cd examples/qemu 
* Run "./01.build.sh".
