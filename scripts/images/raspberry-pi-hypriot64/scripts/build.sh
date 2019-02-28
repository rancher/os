#!/bin/bash
set -e -x

cd $(dirname $0)/..

# create build directory for assembling our image filesystem
mkdir -p build/{boot,root,basefs} dist

cp assets/*.tar.gz build/

#---build SD card image---

# size of root and boot partion (in MByte)
IMAGE_TOTAL_SIZE=2048
BOOT_PARTITION_START=2048
BOOT_PARTITION_SIZE=25
#---don't change here---
BOOT_PARTITION_OFFSET="$((BOOT_PARTITION_START*512))"
BOOT_PARTITION_BYTES="$((BOOT_PARTITION_SIZE*1024*1024))"
BOOT_PARTITION_SECTORS="$((BOOT_PARTITION_SIZE*1024*2))"
ROOT_PARTITION_START="$((BOOT_PARTITION_START+BOOT_PARTITION_SECTORS+1))"
ROOT_PARTITION_OFFSET="$((ROOT_PARTITION_START*512))"
#---don't change here---

# create image file with two partitions (FAT32, EXT4)
dd if=/dev/zero of=build/run.img bs=1MiB count=$IMAGE_TOTAL_SIZE
echo -e "o\nn\np\n1\n${BOOT_PARTITION_START}\n+${BOOT_PARTITION_SECTORS}\nt\nc\nn\np\n2\n${ROOT_PARTITION_START}\n\nw\n" | fdisk build/run.img
fdisk -l build/run.img
ls -al build/run.img

# partition #1 - Type= c W95 FAT32 (LBA)
losetup
losetup -f
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
mkdir -p build/basefs
tar -C build/basefs -zxvf build/kernel.tar.gz
tar -C build/basefs -zxvf build/rpi-bootfiles.tar.gz

# populate kernel, bootloader and RancherOS rootfs
cp -R build/basefs/* build/root
tar -xf assets/rootfs_arm64.tar.gz -C build/root
echo "+dwc_otg.lpm_enable=0 console=tty1 root=/dev/mmcblk0p2 rootfstype=ext4 cgroup-enable=memory swapaccount=1 elevator=deadline rootwait console=ttyAMA0,115200 console=tty0 rancher.password=rancher rancher.autologin=ttyAMA0 rw init=/init" > build/root/boot/cmdline.txt
# enable serial console mode for rpi3
echo "enable_uart=1" > build/root/boot/config.txt

## wireless support
mkdir -p build/root/lib/firmware/brcm
pushd build/root/lib/firmware/brcm
curl -sL -o brcmfmac43430-sdio.txt https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43430-sdio.txt
curl -sL -o brcmfmac43430-sdio.bin https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43430-sdio.bin
curl -sL -o brcmfmac43455-sdio.bin https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43455-sdio.bin
curl -sL -o brcmfmac43455-sdio.clm_blob https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43455-sdio.clm_blob
curl -sL -o brcmfmac43455-sdio.txt https://raw.githubusercontent.com/RPi-Distro/firmware-nonfree/master/brcm/brcmfmac43455-sdio.txt
popd

# TODO: we need to remove these lines
# mitigate this issue: https://github.com/rancher/os/issues/2674
pushd build/root/usr/share/ros/
sed -i 's/io.docker.compose.rebuild: always/io.docker.compose.rebuild\: "false"/g' os-config.yml
popd

# show details
tree -a -L 3 build/root
df -h

# unmount partitions (loopback devices will be removed automatically)
umount build/root/boot
umount build/root

# package, compress and export image file
mv build/run.img build/rancheros-raspberry-pi64.img
zip dist/rancheros-raspberry-pi64.zip build/rancheros-raspberry-pi64.img
ls -alh dist

# cleanup build environment
rm -fr build
