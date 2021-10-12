FROM opensuse/leap:15.3 AS build
RUN zypper ref
RUN zypper in -y squashfs xorriso go1.16 upx busybox-static curl
RUN curl -Lo /usr/bin/luet https://github.com/mudler/luet/releases/download/0.18.1/luet-0.18.1-linux-$(go env GOARCH) && \
    chmod +x /usr/bin/luet
COPY go.mod go.sum /usr/src/
COPY cmd /usr/src/cmd
COPY pkg /usr/src/pkg
RUN cd /usr/src && \
    CGO_ENABLED=0 go build -ldflags "-extldflags -static -s" -o /usr/sbin/ros-installer ./cmd/ros-installer && \
    upx /usr/sbin/ros-installer

FROM scratch AS framework
COPY --from=build /usr/bin/busybox-static /usr/bin/busybox
COPY --from=build /usr/bin/luet /usr/bin/luet
COPY framework/files/etc/luet/luet.yaml /etc/luet/luet.yaml
COPY --from=build /etc/ssl/certs /etc/ssl/certs

ARG CACHEBUST
ENV LUET_NOLOCK=true
RUN ["luet", \
    "install", "--no-spinner", "-d", "-y", \
    "selinux/k3s", \
    "selinux/rancher", \
    "meta/cos-minimal", \
    "utils/k9s", \
    "utils/rancherd", \
    "utils/nerdctl"]

COPY --from=build /usr/sbin/ros-installer /usr/sbin/ros-installer
COPY framework/files/ /
RUN ["/usr/bin/busybox", "sh", "-c", "if [ -e /etc/luet/luet.yaml.$(busybox uname -m) ]; then busybox mv -f /etc/luet/luet.yaml.$(busybox uname -m) /etc/luet/luet.yaml; fi && busybox rm -f /etc/luet/luet.yaml.*"]
RUN ["/usr/bin/busybox", "rm", "-rf", "/var", "/etc/ssl", "/usr/bin/busybox"]

# Make OS image
FROM opensuse/leap:15.3 as os
RUN zypper ref
RUN zypper in -y \
    apparmor-parser \
    avahi \
    bash-completion \
    conntrack-tools \
    coreutils \
    curl \
    device-mapper \
    dmidecode \
    dosfstools \
    dracut \
    e2fsprogs \
    ethtool \
    findutils \
    gawk \
    gptfdisk \
    grub2-i386-pc \
    grub2-x86_64-efi \
    haveged \
    hdparm \
    iotop \
    iproute2 \
    iptables \
    iputils \
    issue-generator \
    jq \
    kernel-default \
    kernel-firmware-bnx2 \
    kernel-firmware-chelsio \
    kernel-firmware-i915 \
    kernel-firmware-intel \
    kernel-firmware-iwlwifi \
    kernel-firmware-liquidio \
    kernel-firmware-marvell \
    kernel-firmware-mediatek \
    kernel-firmware-mellanox \
    kernel-firmware-network \
    kernel-firmware-platform \
    kernel-firmware-qlogic \
    kernel-firmware-realtek \
    kernel-firmware-usb-network \
    less \
    lshw \
    lsof \
    lsscsi \
    lvm2 \
    mdadm \
    multipath-tools \
    nano \
    netcat-openbsd \
    nfs-utils \
    open-iscsi \
    open-vm-tools \
    parted \
    pciutils \
    pigz \
    policycoreutils \
    procps \
    psmisc \
    python-azure-agent \
    qemu-guest-agent \
    rng-tools \
    rsync \
    squashfs \
    strace \
    SUSEConnect \
    sysstat \
    systemd \
    systemd-sysvinit \
    tar \
    tcpdump \
    timezone \
    vim \
    which \
    zstd

# Copy in some local OS customizations
COPY opensuse/files /

# Starting from here are the lines needed for RancherOS to work

# IMPORTANT: Setup rancheros-release used for versioning/upgrade. The
# values here should reflect the tag of the image being built
ARG IMAGE_REPO=norepo
ARG IMAGE_TAG=latest
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /usr/lib/rancheros-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /usr/lib/rancheros-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /usr/lib/rancheros-release

# Copy in framework runtime
COPY --from=framework / /

# Rebuild initrd to setup dracut with the boot configurations
RUN mkinitrd && \
    # aarch64 has an uncompressed kernel so we need to link it to vmlinuz
    kernel=$(ls /boot/Image-* | head -n1) && \
    if [ -e "$kernel" ]; then ln -sf "${kernel#/boot/}" /boot/vmlinuz; fi
