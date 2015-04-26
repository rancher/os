#!/bin/bash

set -ex

if [ -n "$1" ]; then
    # Test for our magic string (it means that the disk was made by ./boot2docker init)
    HEADER=`dd if=$1 bs=1 count=${#MAGIC} 2>/dev/null`

    if [ "$HEADER" = "$MAGIC" ]; then
        # save the preload userdata.tar file
	dd if=$1 of=/userdata.tar bs=1 count=8192
    fi

    mkfs.ext4 -L RANCHER_STATE $1

    if [ -e "/userdata.tar" ]; then
        mount -t ext4 $1 /var/
        mkdir -p /var/lib/rancher/conf/cloud-config.d
        echo $(tar -xvf /userdata.tar)
        AUTHORIZED_KEYS1=$(cat /.ssh/authorized_keys)
        echo -e "#cloud-config\n\nssh_authorized_keys:\n - $AUTHORIZED_KEYS1" > /var/lib/rancher/conf/cloud-config.d/machine.yml
        AUTHORIZED_KEYS2=$(cat /.ssh/authorized_keys2)
        echo -e " - $AUTHORIZED_KEYS2" >> /var/lib/rancher/conf/cloud-config.d/machine.yml
    fi
fi

