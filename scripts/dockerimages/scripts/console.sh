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

RANCHER_HOME=/home/rancher
if [ ! -d ${RANCHER_HOME} ]; then
    mkdir -p ${RANCHER_HOME}
    chown rancher:rancher ${RANCHER_HOME}
    chmod 2755 ${RANCHER_HOME}
fi

for i in $(</proc/cmdline); do
    case $i in
        rancher.password=*)
            PASSWORD=$(echo $i | sed 's/rancher.password=//')
            ;;
    esac
done

if [ -n "$PASSWORD" ]; then
    echo "rancher:$PASSWORD" | chpasswd
fi

cloud-init -execute

if [ -x /var/lib/rancher/conf/cloud-config-script ]; then
    echo "Running /var/lib/rancher/conf/cloud-config-script"
    /var/lib/rancher/conf/cloud-config-script || true
fi

setup_ssh

VERSION="$(rancherctl -v | awk '{print $NF}')"
cat > /etc/lsb-release << EOF
DISTRIB_ID=RancherOS
DISTRIB_RELEASE=${VERSION}
DISTRIB_DESCRIPTION="RancherOS ${VERSION}"
EOF

cat > /etc/respawn.conf << EOF
/sbin/getty 115200 tty1
/sbin/getty 115200 tty2
/sbin/getty 115200 tty3
/sbin/getty 115200 tty4
/sbin/getty 115200 tty5
/sbin/getty 115200 tty6
/usr/sbin/sshd -D
EOF

if ! grep -q '^UseDNS no' /etc/ssh/sshd_config; then
    echo "UseDNS no" >> /etc/ssh/sshd_config
fi

ID_TYPE="busybox"
if [ -e /etc/os-release ] && grep -q 'ID_LIKE=' /etc/os-release; then
    ID_TYPE=$(grep 'ID_LIKE=' /etc/os-release | cut -d'=' -f2)
fi

cat > /etc/os-release << EOF
NAME="RancherOS"
VERSION=$VERSION
ID=rancheros
ID_LIKE=$ID_TYPE
VERSION_ID=$VERSION
PRETTY_NAME="RancherOS"
HOME_URL=
SUPPORT_URL=
BUG_REPORT_URL=
BUILD_ID=
EOF

if ! grep -q "$(hostname)" /etc/hosts; then
    echo 127.0.1.1 $(hostname) >> /etc/hosts
fi

if [ -x /opt/rancher/bin/start.sh ]; then
    echo Executing custom script
    /opt/rancher/bin/start.sh || true
fi

exec respawn -f /etc/respawn.conf
