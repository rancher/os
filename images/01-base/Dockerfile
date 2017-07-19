FROM rancher/strato-min:0.0.2
# FROM amd64=rancher/strato-min:0.0.2 arm64=rancher/strato-min:0.0.2_arm64 arm=skip


# You can build a locel strato repository, and run `make webserver` in that repo
# and make build-images will automatically change the build-arg
ARG STRATO_BIN=https://github.com/rancher/strato/releases/download/0.0.2
ARG STRATO_REPO=https://github.com/rancher/strato-packages/raw/master/0.0.2/amd64/

RUN wget -O /sbin/strato "${STRATO_BIN}/strato_$(get-arch)" \
    && chmod +x /sbin/strato \
    && strato --source="${STRATO_REPO}" add bash \
    openssh \
    zlib \
    e2fsprogs \
    agetty \
    partx \
    sfdisk \
    libuuid \
    libblkid \
    kmod \
    xz \
    iproute2 \
    dhcpcd \
    eudev \
    jq \
    acpi \
    rsync \
    libestr \
    json-c \
    rsyslog \
    sudo \
    parted \
    ntp \
    iptables \
    && rm /sbin/strato

RUN mkdir -p /etc/acpi
RUN rm -rf /etc/ssh/*key*
RUN sed -i s/"\/bin\/ash"/"\/bin\/bash"/g /etc/passwd \
    && rm /bin/sh && ln -s /bin/bash /bin/sh

RUN addgroup -g 1100 rancher && \
    addgroup -g 1101 docker && \
    addgroup -g 1103 sudo && \
    adduser -u 1100 -G rancher -D -h /home/rancher -s /bin/bash rancher && \
    adduser -u 1101 -G docker -D -h /home/docker -s /bin/bash docker && \
    adduser rancher docker && \
    adduser rancher sudo && \
    adduser rancher input && \
    adduser docker sudo && \
    adduser docker input && \
    echo '%sudo ALL=(ALL) ALL' >> /etc/sudoers

COPY inputrc /etc/inputrc
COPY growpart /usr/bin/growpart

# TODO: rebase this properly
RUN sed -i s/"partx --update \"\$part\" \"\$dev\""/"partx --update --nr \"\$part\" \"\$dev\""/g /usr/bin/growpart && \
    sed -i -e 's/duid/clientid/g' /etc/dhcpcd.conf

ENTRYPOINT ["/usr/bin/ros", "entrypoint"]
