#!/bin/bash
set -e -x

BOOT_DEVICE=/dev/sda1
DEVICE=/dev/sda2

cd $(dirname $0)/..

mkdir -p dist build

cp assets/hypriot*.img build/run.img
guestfish --progress-bars --verbose -a build/run.img run \
    : mount $DEVICE / \
    : tar-out /lib/modules modules.tar \
    : tar-out /lib/firmware firmware.tar \
    : umount / \
    : zero-device $DEVICE \
    : mkfs ext4 $DEVICE \
    : mount $DEVICE / \
    : mkdir-p /lib/modules \
    : mkdir-p /lib/firmware \
    : set-verbose false \
    : echo tar-in modules.tar /lib/modules \
    : tar-in modules.tar /lib/modules \
    : echo tar-in firmware.tar /lib/firmware \
    : tar-in firmware.tar /lib/firmware \
    : echo tgz-in assets/rootfs_arm.tar.gz / \
    : tgz-in assets/rootfs_arm.tar.gz / \
    : set-verbose true \
    : umount / \
    : mount $BOOT_DEVICE / \
    : write /cmdline.txt "+dwc_otg.lpm_enable=0 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 cgroup-enable=memory swapaccount=1 elevator=deadline rootwait console=ttyAMA0,115200 kgdboc=ttyAMA0,115200 console=tty0 rancher.password=rancher rw init=/init
" \
    : umount /

zip dist/rancheros-rpi2.zip build/run.img
