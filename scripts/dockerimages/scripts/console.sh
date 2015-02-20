#!/bin/sh
# Test

CLOUD_CONFIG_FILE=/var/lib/rancher/cloud-config

if [ -s $CLOUD_CONFIG_FILE ]; then
	cloud-init --from-file $CLOUD_CONFIG_FILE
fi

RANCER_HOME=/home/rancher
if [ ! -d ${RANCER_HOME} ]; then
    mkdir -p ${RANCER_HOME}
    chown rancher:rancher ${RANCER_HOME}
    chmod 2755 ${RANCER_HOME}
fi

cat > /etc/respawn.conf << EOF
/sbin/getty 115200 tty1
/sbin/getty 115200 tty2
/sbin/getty 115200 tty3
/sbin/getty 115200 tty4
/sbin/getty 115200 tty5
/sbin/getty 115200 tty6
EOF

exec respawn -f /etc/respawn.conf
