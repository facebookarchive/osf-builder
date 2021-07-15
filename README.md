Open System Firmware top level build system

This OSF top level build system (osf) repo is used for the MP-NDA engineering
collaboration by AWS/ByteDance/FB/Google and Intel under Intel multi-party
permission letter. 

This repo supports getting code base and building for:
+--------------+----------+--------+-------------+-------------+
| Platform     |  Vendor  | Socket |Board Status |OSF Status   |
|--------------+----------+--------+-------------+-------------+
|ArcherCity CRB|  Intel   | dual   |C steppings  | Alpha       |
|CraterLake    |  OCP     | single |pre-power on |pre-power on |
| S9S          |  Quanta  | dual   |  EVT        |power on     |
+--------------+----------+--------+-------------+-------------+

This repo has 2 components:
a. scripts. This build script (build.sh) clones osf-builder open source OSF
** Patches: Patch files to be applied to kernel, u-root and coreboot.
** Resources: Local tar balls and static binaries to be included in initramfs.
build system, and executes osf-builder build script with platform metadata
provided, in the form of environment variables. 
b. Platform metadata.
** configs directory: code base is defined here, in JSON format.
** artifacts directory: kernel, coreboot configurations are defined here.
** patches directory: kernel, u-root, coreboot patches are stored in this
directory.
** resources: static binaries to be included in initramfs image are stored.

Following are build server dependencies:
* Compilers and build tools - These include jq, vpd tools, go compilers.
* gcc version needs to be greater than or equal to 5.0.0. 

Following are general guidance for the MP-NDA collaboration:
* kernel and u-root shall be upstream only.
* coreboot shall be coreboot-spr-sp repo only.
* patches are needed sometimes in this osf repo, but the number should be
limited, and they should be upstreamed as soon as possible.

## How to build
* Clone the repo.
* Run "PLATFORM=<platform name> ./build.sh". If PLATFORM is not defined,
by default ArcherCity CRB image is built.
