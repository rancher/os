FROM strato-min
# FROM amd64=stratopkg/strato-min:0.0.1 arm64=stratopkg/strato-min:0.0.1_arm64 arm=skip

RUN wget -O /sbin/strato "https://github.com/joshwget/strato/releases/download/0.0.1/strato_$(get-arch)" \
    && chmod +x /sbin/strato \
    && strato add bash \
    openssh \
    zlib \
    e2fsprogs \
    util-linux#agetty \
    util-linux#partx \
    util-linux#libuuid \
    util-linux#libblkid \
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
    && rm /sbin/strato

RUN mkdir -p /etc/acpi
RUN rm -rf /etc/ssh/*key*

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
