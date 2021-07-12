ARG LUET_VERSION=0.16.7
FROM quay.io/luet/base:$LUET_VERSION AS luet

FROM registry.suse.com/suse/sle15:15.3 AS base

# Copy luet from the official images
COPY --from=luet /usr/bin/luet /usr/bin/luet

ARG ARCH=amd64
ENV ARCH=${ARCH}
RUN zypper rm -y container-suseconnect
RUN zypper ar --priority=200 http://download.opensuse.org/distribution/leap/15.3/repo/oss repo-oss
RUN zypper --no-gpg-checks ref
COPY files/etc/luet/luet.yaml /etc/luet/luet.yaml

FROM base as tools
ENV LUET_NOLOCK=true
RUN zypper in -y squashfs xorriso
COPY tools /
RUN luet install -y toolchain/luet-makeiso

FROM base
RUN zypper in -y \
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
    nfs-utils \
    open-iscsi \
    open-vm-tools \
    parted \
    pigz \
    policycoreutils \
    procps \
    python-azure-agent \
    qemu-guest-agent \
    rng-tools \
    rsync \
    squashfs \
    strace \
    systemd \
    systemd-sysvinit \
    tar \
    timezone \
    vim \
    which

ARG CACHEBUST
RUN luet install -y \
    toolchain/yip \
    utils/installer \
    system/cloud-config \
    system/cos-setup \
    system/immutable-rootfs \
    system/grub-config \
    selinux/k3s \
    selinux/rancher \
    utils/k9s \
    utils/nerdctl \
    utils/rancherd

COPY files/ /
RUN mkinitrd

ARG OS_NAME=RancherOS
ARG OS_VERSION=999
ARG OS_GIT=dirty
ARG OS_REPO=norepo/norepo
ARG OS_LABEL=latest
RUN envsubst >/usr/lib/os-release </usr/lib/os-release.tmpl && \
    rm /usr/lib/os-release.tmpl
