#!/bin/bash
set -e

setup_ssh()
{
    for i in rsa dsa ecdsa ed25519; do
        local output=/etc/ssh/ssh_host_${i}_key
        if [ ! -e $output ]; then
            local saved="$(rancherctl config get ssh.keys.${i})"
            local pub="$(rancherctl config get ssh.keys.${i}-pub)"

            if [[ -n "$saved" && -n "$pub" ]]; then
                (
                    umask 477
                    echo "$saved" > ${output}
                    echo "$pub" > ${output}.pub
                )
            else
                ssh-keygen -f $output -N '' -t $i
                rancherctl config set -- ssh.keys.${i} "$(<${output})"
                rancherctl config set -- ssh.keys.${i}-pub "$(<${output}.pub)"
            fi
        fi
    done

    mkdir -p /var/run/sshd
}

cloud-init -execute

setup_ssh

cat > /etc/respawn.conf << EOF
/sbin/getty 115200 tty1
/sbin/getty 115200 tty2
/sbin/getty 115200 tty3
/sbin/getty 115200 tty4
/sbin/getty 115200 tty5
/sbin/getty 115200 tty6
/usr/sbin/sshd -D
EOF

RANCHER_HOME=/home/rancher
if [ ! -d ${RANCHER_HOME} ]; then
    mkdir -p ${RANCHER_HOME}
    chown rancher:rancher ${RANCHER_HOME}
    chmod 2755 ${RANCHER_HOME}
fi

if ! grep -q "$(hostname)" /etc/hosts; then
    echo 127.0.1.1 $(hostname) >> /etc/hosts
fi

if [ -x /opt/rancher/bin/start.sh ]; then
    /opt/rancher/bin/start.sh
fi

exec respawn -f /etc/respawn.conf
