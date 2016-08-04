FROM alpine:3.4
RUN apk -U add \
    bash \
    e2fsprogs \
    parted \
    sfdisk \
    openssh \
    xz \
    iproute2 \
    dhcpcd \
    eudev \
    jq \
    acpi \
    rsync \
    rsyslog \
    sudo \
    openntpd && \
    rm -rf /var/lib/apk/*
RUN addgroup -g 1100 rancher && \
    addgroup -g 1101 docker && \
    addgroup -g 1103 sudo && \
    addgroup input && \
    adduser -u 1100 -G rancher -D -h /home/rancher -s /bin/bash rancher && \
    adduser -u 1101 -G docker -D -h /home/docker -s /bin/bash docker && \
    adduser rancher docker && \
    adduser rancher sudo && \
    adduser rancher input && \
    adduser docker sudo && \
    adduser docker input && \
    echo '%sudo ALL=(ALL) ALL' >> /etc/sudoers

RUN rm -f /sbin/poweroff /sbin/shutdown /sbin/reboot /sbin/halt /usr/sbin/poweroff /usr/sbin/shutdown /usr/sbin/reboot /usr/sbin/halt /etc/motd
RUN mkdir -p /etc/acpi

COPY inputrc /etc/inputrc
COPY entry.sh /usr/sbin/entry.sh
COPY growpart /usr/bin/growpart
RUN sed -i -e 's/duid/clientid/g' /etc/dhcpcd.conf

ENTRYPOINT ["/usr/sbin/entry.sh"]
