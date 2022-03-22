FROM burmilla/os-rootfs

COPY . /

RUN rm /sbin/poweroff /sbin/reboot /sbin/halt && \
    sed -i '/^root/s!/bin/sh!/bin/bash!' /etc/passwd && \
    echo 'BurmillaOS \n \l' > /etc/issue && \
    rm -rf /run \
       /linuxrc \
       /etc/os-release \
       /var/cache \
       /var/lock \
       /var/log \
       /var/run \
       /var/spool \
       /var/lib/misc && \
    mkdir -p \
       /home \
       /run \
       /var/cache \
       /var/lock \
       /var/log \
       /var/run \
       /var/spool && \
    passwd -l root && \
    addgroup -g 1100 rancher && \
    addgroup -g 1101 docker && \
    adduser -u 1100 -G rancher -D -h /home/rancher -s /bin/bash rancher && \
    adduser -u 1101 -G docker -D -h /home/docker -s /bin/bash docker && \
    adduser rancher docker && \
    adduser rancher sudo && \
    adduser docker sudo && \
    echo '%sudo ALL=(ALL) ALL' >> /etc/sudoers && \
    ln -s /dev/null /etc/udev/rules.d/80-net-name-slot.rules && \
    sed -i s/"partx --update \"\$part\" \"\$dev\""/"partx --update --nr \"\$part\" \"\$dev\""/g /usr/bin/growpart && \
    sed -i -e 's/duid/clientid/g' /etc/dhcpcd.conf && \
    echo -e '\n# Ignore Docker container interfaces\ndenyinterfaces veth*\n' >> /etc/dhcpcd.conf && \
    rm -f /etc/wpa_supplicant.conf && \
    ln -s /usr/share/dhcpcd/hooks/10-wpa_supplicant /lib/dhcpcd/dhcpcd-hooks/ && \
    rm -f /usr/share/bash-completion/completions/* && \
    chmod 555 /lib/dhcpcd/dhcpcd-run-hooks && \
    sed -i 1,10d /etc/rsyslog.conf && \
    echo "*.*                /var/log/syslog" >> /etc/rsyslog.conf && \
    \
    addgroup -g 1200 user-docker && \
    adduser -u 1200 -G user-docker -S -H user-docker && \
    echo 'user-docker:100000:65536' > /etc/subuid && \
    echo 'user-docker:100000:65536' > /etc/subgid
# dump kernel log to console (but after we've finished booting)
#    echo "kern.*                /dev/console" >> /etc/rsyslog.conf

ENTRYPOINT ["/usr/bin/ros", "entrypoint"]
