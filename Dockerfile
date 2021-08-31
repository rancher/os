FROM opensuse/leap:15.3 AS build
RUN zypper in -y squashfs xorriso go1.16 upx busybox-static
COPY go.mod go.sum /usr/src/
COPY cmd /usr/src/cmd
COPY pkg /usr/src/pkg
RUN cd /usr/src && \
    CGO_ENABLED=0 go build -ldflags "-extldflags -static -s" -o /usr/sbin/ros-installer ./cmd/ros-installer && \
    upx /usr/sbin/ros-installer

FROM scratch AS framework
COPY --from=build /usr/bin/busybox-static /usr/bin/busybox
COPY --from=quay.io/luet/base:0.17.8 /usr/bin/luet /usr/bin/luet
COPY framework/files/etc/luet/luet.yaml /etc/luet/luet.yaml
COPY --from=build /etc/ssl/certs /etc/ssl/certs

ARG CACHEBUST
ENV LUET_NOLOCK=true
RUN ["luet", \
    "install", "--no-spinner", "-d", "-y", \
    "selinux/k3s", \
    "selinux/rancher", \
    "system/base-dracut-modules", \
    "system/cloud-config", \
    "system/cos-setup", \
    "system/grub2-config", \
    "system/immutable-rootfs", \
    "toolchain/yip", \
    "utils/installer", \
    "utils/k9s", \
    "utils/nerdctl"]

COPY --from=build /usr/sbin/ros-installer /usr/sbin/ros-installer
COPY framework/files/ /
RUN ["/usr/bin/busybox", "rm", "-rf", "/var", "/etc/ssl", "/usr/bin/busybox"]

# Make OS image
FROM opensuse/leap:15.3 as os
RUN zypper in -y \
    avahi \
    bash-completion \
    conntrack-tools \
    coreutils \
    curl \
    device-mapper \
    dosfstools \
    dracut \
    e2fsprogs \
    findutils \
    gawk \
    gptfdisk \
    grub2-i386-pc \
    grub2-x86_64-efi \
    haveged \
    iproute2 \
    iptables \
    iputils \
    issue-generator \
    jq \
    kernel-default \
    kernel-firmware-bnx2 \
    kernel-firmware-i915 \
    kernel-firmware-intel \
    kernel-firmware-iwlwifi \
    kernel-firmware-mellanox \
    kernel-firmware-network \
    kernel-firmware-platform \
    kernel-firmware-realtek \
    less \
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
    pigz \
    policycoreutils \
    psmisc \
    procps \
    python-azure-agent \
    qemu-guest-agent \
    rsync \
    squashfs \
    strace \
    SUSEConnect \
    systemd \
    systemd-sysvinit \
    tcpdump \
    tar \
    timezone \
    vim \
    which

# Copy in some local OS customizations
COPY opensuse/files /

# Starting from here are the lines needed for RancherOS to work

# IMPORTANT: Setup rancheros-release used for versioning/upgrade. The
# values here should reflect the tag of the image building built
ARG IMAGE_REPO=norepo
ARG IMAGE_TAG=latest
RUN echo "IMAGE_REPO=${IMAGE_REPO}"          > /usr/lib/rancheros-release && \
    echo "IMAGE_TAG=${IMAGE_TAG}"           >> /usr/lib/rancheros-release && \
    echo "IMAGE=${IMAGE_REPO}:${IMAGE_TAG}" >> /usr/lib/rancheros-release

# Copy in framework runtime
COPY --from=framework / /

# Rebuild initrd to setup dracut with the boot configurations
RUN mkinitrd
