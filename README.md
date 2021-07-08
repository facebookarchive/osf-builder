osf-builder

This osf-builder repo is an open source repo to faciliate the development community to collaborate on Open System Firmware (OSF).

The OSF boot process has following phases:
- Coreboot. This component initize SoC and platform component (such as memory) to the extent that kernel on flash can be loaded. In future, UEFI support may be added.
- Kernel. The kernel binary is built with Linux kernel code using Linuxboot specific Kconfig, which is needed to reduce kernel size.
- Initramfs. The initramfs is u-root, which is similar to busybox. It is built with go language.

The osf-builder repo has 3 components:
a. A code syncer. It is in cmd directory. It has a bunch of go source code. It builds into getdpes binary, which is able to sync down code based on JSON format configuration. The code may be obtained from a certain commit of a certain branch of a certain repo, or from a local compressed tar ball, or from a URL.
b. Build scripts. The build scripts get source code, build initramfs, kernel and then coreboot.
c. A sample project. The examples/qemu directory contains a sample project. A project may have following components:
** Configs: The code base configuration in JSON format.
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
