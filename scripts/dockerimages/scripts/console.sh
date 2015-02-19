#!/bin/sh

#if [ -t 1 ]; then
    #exec /bin/sh
#else

CLOUD_CONFIG_FILE=/var/lib/rancher/cloud-config

if [ -s $CLOUD_CONFIG_FILE ]; then
	cloudinit --from-file $CLOUD_CONFIG_FILE
fi
    exec respawn << EOF
/sbin/getty 115200 tty1
/sbin/getty 115200 tty2
/sbin/getty 115200 tty3
/sbin/getty 115200 tty4
/sbin/getty 115200 tty5
/sbin/getty 115200 tty6
EOF
#fi
