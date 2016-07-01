#!/bin/bash
set -e -x

cd $(dirname $0)/..

# create build directory for assembling our image filesystem
mkdir -p build/{boot,root,basefs} dist

cp assets/raspberrypi-bootloader_*_armhf.deb build/bootloader.deb
cp assets/rpi3-bootfiles.tar.gz build/rpi3-bootfiles.tar.gz

#---build SD card image---

# size of root and boot partion (in MByte)
IMAGE_TOTAL_SIZE=500
BOOT_PARTITION_START=2048
BOOT_PARTITION_SIZE=25
#---don't change here---
BOOT_PARTITION_OFFSET="$((BOOT_PARTITION_START*512))"
BOOT_PARTITION_BYTES="$((BOOT_PARTITION_SIZE*1024*1024))"
BOOT_PARTITION_SECTORS="$((BOOT_PARTITION_SIZE*1024*2))"
ROOT_PARTITION_START="$((BOOT_PARTITION_START+BOOT_PARTITION_SECTORS))"
ROOT_PARTITION_OFFSET="$((ROOT_PARTITION_START*512))"
#---don't change here---

# create image file with two partitions (FAT32, EXT4)
dd if=/dev/zero of=build/run.img bs=1MiB count=$IMAGE_TOTAL_SIZE
echo -e "o\nn\np\n1\n${BOOT_PARTITION_START}\n+${BOOT_PARTITION_SECTORS}\nt\nc\nn\np\n2\n${ROOT_PARTITION_START}\n\nw\n" | fdisk build/run.img
fdisk -l build/run.img
ls -al build/run.img

# partition #1 - Type= c W95 FAT32 (LBA)
losetup -d /dev/loop0 || /bin/true
losetup --offset $BOOT_PARTITION_OFFSET --sizelimit $BOOT_PARTITION_BYTES /dev/loop0 build/run.img
mkfs.vfat -n RancherOS /dev/loop0
losetup -d /dev/loop0

# partition #2 - Type=83 Linux
losetup -d /dev/loop1 || /bin/true
losetup --offset $ROOT_PARTITION_OFFSET /dev/loop1 build/run.img
mkfs.ext4 -O ^has_journal -b 4096 -L rootfs /dev/loop1
losetup -d /dev/loop1

# mount partitions as loopback devices
mount -t ext4 -o loop=/dev/loop1,offset=$ROOT_PARTITION_OFFSET build/run.img build/root
mkdir -p build/root/boot
mount -t vfat -o loop=/dev/loop0,offset=$BOOT_PARTITION_OFFSET build/run.img build/root/boot
echo "RancherOS: boot partition" > build/root/boot/boot.txt
echo "RancherOS: root partition" > build/root/root.txt

# unpack and cleanup the basefs
#- doing this on a local folder keeps our resulting image clean (no dirty blocks from a delete)
dpkg-deb -x build/bootloader.deb build/basefs
# upgrade Raspberry Pi bootfile for RPi3 support
tar xvzf build/rpi3-bootfiles.tar.gz -C build/basefs/boot
# remove RPi1 kernel, we only support RPi2 and RPi3 in ARMv7 mode
rm -fr build/basefs/boot/kernel.img
rm -fr build/basefs/lib/modules/{4.1.17+,4.1.17-hypriotos+}

# populate kernel, bootloader and RancherOS rootfs
cp -R build/basefs/* build/root
tar -xf assets/rootfs_arm.tar.gz -C build/root
echo "+dwc_otg.lpm_enable=0 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 cgroup-enable=memory swapaccount=1 elevator=deadline rootwait console=ttyAMA0,115200 kgdboc=ttyAMA0,115200 console=tty0 rancher.password=rancher rw init=/init" > build/root/boot/cmdline.txt

# show details
tree -a -L 3 build/root
df -h

# unmount partitions (loopback devices will be removed automatically)
umount build/root/boot
umount build/root

# package, compress and export image file
mv build/run.img build/rancheros-raspberry-pi.img
zip dist/rancheros-raspberry-pi.zip build/rancheros-raspberry-pi.img
ls -alh dist

# cleanup build environment
rm -fr build
