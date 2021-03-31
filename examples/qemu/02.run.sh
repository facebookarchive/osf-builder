# sudo is required to enable KVM below
sudo qemu-system-x86_64 \
    `# the machine type specified in the coreboot mainboard configuration` \
    -M q35 \
    `# use KVM to avail of hardware virtualization extensions` \
    -enable-kvm \
    `# the coreboot ROM to run as system firmware` \
    -bios coreboot/build/coreboot.rom \
    `# the amount of RAM in MB` \
    -m 1024 \
    `# RNG to avoid DHCP lockups when waiting for entropy` \
    -object rng-random,filename=/dev/urandom,id=rng0 \
    `# redirect all the output to the console` \
    -nographic
