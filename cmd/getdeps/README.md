# getdeps

`getdeps` is a tool to fetch OSF dependencies and put them in the local
directory so that it is ready to build. It helps with component versioning and
multiple configurations, and uses JSON to describe a configuration.

OSF is made by multiple components, for which you may or may not have the full
source code. With `getdeps` you can specify which components you want, along
with their exact versions and locations (e.g. URL or git repository).

## Quickstart

```
cd cmd/getdeps
go build
./getdeps --config ../../configs/qemu-x86_64.json
cat generated_versions.json
```

The above commands will fetch the required components to build OSF for QEmu with
machine type q35 on x86_64. Each component is fetched in an homonym directory in
the current directory where `getdeps` is run.

## Components

Currently getdeps supports the following components:
* `coreboot`: from the [coreboot project](https://coreboot.org), this component
  is currently used to initialize the platform, and to load a LinuxBoot payload,
  made by the `kernel` and `initramfs` components below. It is fetched via git.
* `kernel`: this is based on the Linux kernel. You can specify a git repository
  or a tarball.
* `initramfs`: based on [u-root](https://github.com/u-root/u-root), an embedded
  environment with various bootloaders that run in userspace on the firmware
  kernel.

Each component is controlled by the configuration files described below.

## Configuration files

TODO

## URL overrides

TODO
