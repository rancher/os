#!/bin/bash
set -ex

INSTALLER_IMAGE=rancher/os:v0.7.1

ros config set rancher.network.interfaces.eth1.dhcp false
if grep eth2 /proc/net/dev; then
    ros config set rancher.network.interfaces.eth0.dhcp false
    ros config set rancher.network.interfaces.eth2.dhcp true
    system-docker restart network
fi

for ((i=0;i<30;i++)); do
    if system-docker pull ${INSTALLER_IMAGE}; then
        break
    fi
    sleep 1
done

TINKERBELL_URL=$(cat /proc/cmdline | sed -e 's/^.*tinkerbell=//' -e 's/ .*$//')/phone-home

tinkerbell_post()
{
    system-docker run rancher/curl -X POST -H "Content-Type: application/json" -d "{\"type\":\"provisioning.$1\",\"body\":\"$2\"}" ${TINKERBELL_URL}
}

tinkerbell_post 104 "Connected to magic install system"

DEV_PREFIX=/dev/sd
if [ -e /dev/vda ]; then
    DEV_PREFIX=/dev/vd
fi

BOOT=${DEV_PREFIX}a1
BOOT_TYPE=83
SWAP=${DEV_PREFIX}a5
SWAP_TYPE=82
OEM=${DEV_PREFIX}a6
OEM_TYPE=83
ROOT=${DEV_PREFIX}a7
ROOT_TYPE=83
RAID=false

wait_for_dev()
{
    for DEV; do
        for ((i=0;i<10;i++)); do
            if [ ! -e $DEV ]; then
                partprobe || true
                sleep 1
            else
                break
            fi
        done
    done
}

modprobe md || true

dd if=/dev/zero of=${DEV_PREFIX}a bs=1M count=1

if [ -e ${DEV_PREFIX}b ]; then
    dd if=/dev/zero of=${DEV_PREFIX}b bs=1M count=1
    RAID=true
    BOOT=/dev/md1
    BOOT_TYPE=fd
    SWAP=/dev/md5
    SWAP_TYPE=fd
    OEM=/dev/md6
    OEM_TYPE=fd
    ROOT=/dev/md7
    ROOT_TYPE=fd
fi


# Partition BOOT
echo -e "n\np\n1\n\n+2G\nt\n${BOOT_TYPE}\nw" | fdisk ${DEV_PREFIX}a || true
partprobe || true
# Partition Extended
echo -e "n\ne\n\n\n\nw" | fdisk ${DEV_PREFIX}a || true
partprobe || true
# Partition SWAP
echo -e "n\nl\n\n+2G\n\nt\n5\n${SWAP_TYPE}\nw" | fdisk ${DEV_PREFIX}a || true
partprobe || true
# Partition OEM
echo -e "n\nl\n\n+100M\n\nt\n6\n${OEM_TYPE}\nw" | fdisk ${DEV_PREFIX}a || true
partprobe || true
# Partition ROOT
echo -e "n\nl\n\n\n\nt\n7\n${ROOT_TYPE}\nw" | fdisk ${DEV_PREFIX}a || true
partprobe || true
# Make boot active
echo -e "a\n1\nw" | fdisk ${DEV_PREFIX}a || true
partprobe || true

if [ "$RAID" = "true" ]; then
    sfdisk --dump ${DEV_PREFIX}a | sfdisk --no-reread ${DEV_PREFIX}b

    wait_for_dev ${DEV_PREFIX}b1 ${DEV_PREFIX}b5 ${DEV_PREFIX}b6 ${DEV_PREFIX}b7 ${DEV_PREFIX}a1 ${DEV_PREFIX}a5 ${DEV_PREFIX}a6 ${DEV_PREFIX}a7

    mdadm --create $BOOT --level=1 --metadata=1.0 --raid-devices=2 ${DEV_PREFIX}a1 ${DEV_PREFIX}b1
    mdadm --create $SWAP --level=1 --metadata=1.2 --raid-devices=2 ${DEV_PREFIX}a5 ${DEV_PREFIX}b5
    mdadm --create $OEM --level=1 --metadata=1.2 --raid-devices=2 ${DEV_PREFIX}a6 ${DEV_PREFIX}b6
    mdadm --create $ROOT --level=1 --metadata=1.2 --raid-devices=2 ${DEV_PREFIX}a7 ${DEV_PREFIX}b7
fi

mkswap -L RANCHER_SWAP $SWAP
system-docker run --privileged --entrypoint /bin/bash ${INSTALLER_IMAGE} -c "mkfs.ext4 -L RANCHER_BOOT $BOOT"
system-docker run --privileged --entrypoint /bin/bash ${INSTALLER_IMAGE} -c "mkfs.ext4 -L RANCHER_STATE $ROOT"
system-docker run --privileged --entrypoint /bin/bash ${INSTALLER_IMAGE} -c "mkfs.ext4 -L RANCHER_OEM $OEM"

tinkerbell_post 105 "Server partitions created"

mkdir -p /mnt/oem
mount $(ros dev LABEL=RANCHER_OEM) /mnt/oem
cat > /mnt/oem/oem-config.yml << EOF
#cloud-config
mounts:
- [ LABEL=RANCHER_SWAP, "", swap, "" ]
EOF
umount /mnt/oem

tinkerbell_post 106 "OEM drive configured"

METADATA=$(system-docker run rancher/curl -sL https://metadata.packet.net/metadata)
eval $(echo ${METADATA} | jq -r '.network.addresses[] | select(.address_family == 4 and .public) | "ADDRESS=\(.address)/\(.cidr)\nGATEWAY=\(.gateway)"')
eval $(echo ${METADATA} | jq -r '.network.interfaces[0] | "MAC=\(.mac)"')
NETWORK_ARGS="rancher.network.interfaces.bond0.address=$ADDRESS rancher.network.interfaces.bond0.gateway=$GATEWAY rancher.network.interfaces.mac:${MAC}.bond=bond0"

tinkerbell_post 107 "Network interface configuration fetched from metadata"

COMMON_ARGS="console=ttyS1,115200n8 rancher.autologin=ttyS1 rancher.cloud_init.datasources=[packet] ${NETWORK_ARGS}"
if [ "$RAID" = "true" ]; then
    ros install -f -t raid -i ${INSTALLER_IMAGE} -d ${DEV_PREFIX}a -a "rancher.state.mdadm_scan ${COMMON_ARGS}" --no-reboot
else
    ros install -f -t noformat -i ${INSTALLER_IMAGE} -d ${DEV_PREFIX}a -a "${COMMON_ARGS}" --no-reboot
fi

tinkerbell_post 109 "Installation finished, rebooting server"
reboot
