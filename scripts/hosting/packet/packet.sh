#!/bin/bash
set -e

if ros version >/dev/null 2>&1; then
    exit 0
fi

apt-get install -y jq curl

mkdir -p /boot/ros

URL_BASE=https://releases.rancher.com/os/latest
curl -L $URL_BASE/vmlinuz > /boot/ros/vmlinuz
curl -L $URL_BASE/initrd > /boot/ros/initrd

eval $(curl -sL https://metadata.packet.net/metadata | jq -r '.network.addresses[] | select(.address_family == 4 and .public) | "ADDRESS=\(.address)/\(.cidr)\nGATEWAY=\(.gateway)"')
eval $(curl -sL https://metadata.packet.net/metadata | jq -r '.network.interfaces[0] | "MAC=\(.mac)"')

cat > /etc/default/grub << "EOF"
# If you change this file, run 'update-grub' afterwards to update
# /boot/grub/grub.cfg.
# For full documentation of the options in this file, see:
#   info -f grub -n 'Simple configuration'

GRUB_DEFAULT=ROS
GRUB_HIDDEN_TIMEOUT=15
GRUB_HIDDEN_TIMEOUT_QUIET=false
GRUB_TIMEOUT=10
GRUB_DISTRIBUTOR=`lsb_release -i -s 2> /dev/null || echo Debian`
GRUB_CMDLINE_LINUX_DEFAULT="console=tty0 console=ttyS1,115200n8"
GRUB_CMDLINE_LINUX=""

# Uncomment to enable BadRAM filtering, modify to suit your needs
# This works with Linux (no patch required) and with any kernel that obtains
# the memory map information from GRUB (GNU Mach, kernel of FreeBSD ...)
#GRUB_BADRAM="0x01234567,0xfefefefe,0x89abcdef,0xefefefef"

# Uncomment to disable graphical terminal (grub-pc only)
GRUB_TERMINAL=console

# The resolution used on graphical terminal
# note that you can use only modes which your graphic card supports via VBE
# you can see them in real GRUB with the command `vbeinfo'
#GRUB_GFXMODE=640x480

# Uncomment if you don't want GRUB to pass "root=UUID=xxx" parameter to Linux
#GRUB_DISABLE_LINUX_UUID=true

# Uncomment to disable generation of recovery mode menu entries
#GRUB_DISABLE_RECOVERY="true"

# Uncomment to get a beep at grub start
#GRUB_INIT_TUNE="480 440 1"
GRUB_TERMINAL=serial
GRUB_SERIAL_COMMAND="serial --speed=115200 --unit=1 --word=8 --parity=no --stop=1"
EOF

cat > /etc/grub.d/50ros << EOF
#!/bin/sh
exec tail -n +3 \$0
# This file provides an easy way to add custom menu entries.  Simply type the
# menu entries you want to add after this comment.  Be careful not to change
# the 'exec tail' line above.
menuentry 'ROS' {
        recordfail
        load_video
        insmod gzio
        insmod part_msdos
        insmod part_msdos
        insmod diskfilter
        insmod mdraid1x
        insmod ext2
        linux   /ros/vmlinuz rancher.state.mdadm_scan rancher.state.directory=ros rancher.network.interfaces.bond0.address=$ADDRESS rancher.network.interfaces.bond0.gateway=$GATEWAY rancher.network.interfaces.mac:${MAC}.bond=bond0 rancher.cloud_init.datasources=[packet] rancher.rm_usr console=tty0 console=ttyS1,115200n8
        initrd  /ros/initrd
}
menuentry 'ROS Debug' {
        recordfail
        load_video
        insmod gzio
        insmod part_msdos
        insmod part_msdos
        insmod diskfilter
        insmod mdraid1x
        insmod ext2
        linux   /ros/vmlinuz rancher.state.mdadm_scan rancher.state.directory=ros rancher.network.interfaces.bond0.address=$ADDRESS rancher.network.interfaces.bond0.gateway=$GATEWAY rancher.network.interfaces.mac:${MAC}.bond=bond0 rancher.cloud_init.datasources=[packet] rancher.rm_usr rancher.network.interfaces.eth*.dhcp=false console=tty0 console=ttyS1,115200n8 rancher.debug rancher.log
        initrd  /ros/initrd
}
EOF

chmod +x /etc/grub.d/50ros

update-grub2

tune2fs -L RANCHER_STATE $(df -h / | sed 1d | awk '{print $1}')

reboot
