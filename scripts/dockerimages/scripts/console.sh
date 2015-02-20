#!/bin/sh
# Test

CLOUD_CONFIG_FILE=/var/lib/rancher/cloud-config

if [ -s $CLOUD_CONFIG_FILE ]; then
	cloud-init --from-file $CLOUD_CONFIG_FILE
fi

for i in rsa dsa ecdsa ed25519; do
    OUTPUT=/etc/ssh/ssh_host_${i}_key
    if [ ! -e $OUTPUT ]; then
        ssh-keygen -f $OUTPUT -N '' -t $i
    fi
done

RANCER_HOME=/home/rancher
if [ ! -d ${RANCER_HOME} ]; then
    mkdir -p ${RANCER_HOME}
    chown rancher:rancher ${RANCER_HOME}
    chmod 2755 ${RANCER_HOME}
fi

chown root:rancher /var/run/docker.sock:/var/run/system-docker.sock

cat > /etc/respawn.conf << EOF
/sbin/getty 115200 tty1
/sbin/getty 115200 tty2
/sbin/getty 115200 tty3
/sbin/getty 115200 tty4
/sbin/getty 115200 tty5
/sbin/getty 115200 tty6
/usr/sbin/sshd -D
EOF

exec respawn -f /etc/respawn.conf
