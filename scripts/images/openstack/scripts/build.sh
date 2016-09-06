#!/bin/bash
set -e

cd $(dirname $0)/..

mkdir -p {dist,build/openstack/latest}

cat > build/openstack/latest/user_data << EOF
#!/bin/bash
set -e

trap "poweroff" EXIT

mount -t 9p -o trans=virtio,version=9p2000.L config-2 /mnt

touch log
sleep 5
openvt -s -- tail -f log &
ros install -d /dev/vda -f --no-reboot >log 2>&1

touch /mnt/success
EOF

rm -f build/{success,hd.img}
qemu-img create -f qcow2 build/hd.img 8G
kvm -curses \
    -drive if=virtio,file=build/hd.img \
    -cdrom assets/rancheros.iso \
    -m 2048 \
    -fsdev local,id=conf,security_model=none,path=$(pwd)/build \
    -device virtio-9p-pci,fsdev=conf,mount_tag=config-2

[ -f build/success ]

echo Converting dist/rancheros-openstack.img
qemu-img convert -c -O qcow2 build/hd.img dist/rancheros-openstack.img
