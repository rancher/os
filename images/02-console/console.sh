#!/bin/bash
set -e -x

setup_ssh()
{
    for i in rsa dsa ecdsa ed25519; do
        local output=/etc/ssh/ssh_host_${i}_key
        if [ ! -s $output ]; then
            local saved="$(ros config get rancher.ssh.keys.${i})"
            local pub="$(ros config get rancher.ssh.keys.${i}-pub)"

            if [[ -n "$saved" && -n "$pub" ]]; then
                (
                    umask 077
                    temp_file=$(mktemp)
                    echo "$saved" > ${temp_file}
                    mv ${temp_file} ${output}
                    temp_file=$(mktemp)
                    echo "$pub" > ${temp_file}
                    mv ${temp_file} ${output}.pub
                )
            else
                ssh-keygen -f $output -N '' -t $i
                ros config set -- rancher.ssh.keys.${i} "$(<${output})"
                ros config set -- rancher.ssh.keys.${i}-pub "$(<${output}.pub)"
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

DOCKER_HOME=/home/docker
if [ ! -d ${DOCKER_HOME} ]; then
    mkdir -p ${DOCKER_HOME}
    chown docker:docker ${DOCKER_HOME}
    chmod 2755 ${DOCKER_HOME}
fi

echo 1000000000 > /proc/sys/fs/file-max

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

setup_ssh

cat > /etc/respawn.conf << EOF
/sbin/getty 115200 tty6
/sbin/getty 115200 tty5
/sbin/getty 115200 tty4
/sbin/getty 115200 tty3
/sbin/getty 115200 tty2
/sbin/getty 115200 tty1
/usr/sbin/sshd -D
EOF

for i in ttyS{0..4} ttyAMA0; do
    if grep -q 'console='$i /proc/cmdline; then
        echo '/sbin/getty 115200' $i >> /etc/respawn.conf
    fi
done

if ! grep -q '^UseDNS no' /etc/ssh/sshd_config; then
    echo "UseDNS no" >> /etc/ssh/sshd_config
fi

if ! grep -q '^PermitRootLogin no' /etc/ssh/sshd_config; then
    echo "PermitRootLogin no" >> /etc/ssh/sshd_config
fi

if ! grep -q '^ServerKeyBits 2048' /etc/ssh/sshd_config; then
    echo "ServerKeyBits 2048" >> /etc/ssh/sshd_config
fi

if ! grep -q '^AllowGroups docker' /etc/ssh/sshd_config; then
    echo "AllowGroups docker" >> /etc/ssh/sshd_config
fi

VERSION="$(ros os version)"
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

echo 'RancherOS \n \l' > /etc/issue
echo $(/sbin/ifconfig | grep -B1 "inet addr" |awk '{ if ( $1 == "inet" ) { print $2 } else if ( $2 == "Link" ) { printf "%s:" ,$1 } }' |awk -F: '{ print $1 ": " $3}') >> /etc/issue

cloud-init -execute

if [ -x /var/lib/rancher/conf/cloud-config-script ]; then
    echo "Running /var/lib/rancher/conf/cloud-config-script"
    /var/lib/rancher/conf/cloud-config-script || true
fi

if [ -x /opt/rancher/bin/start.sh ]; then
    echo Executing custom script
    /opt/rancher/bin/start.sh || true
fi

touch /run/console-done

if [ -x /etc/rc.local ]; then
    echo Executing rc.local
    /etc/rc.local || true
fi

export TERM=linux
exec respawn -f /etc/respawn.conf
