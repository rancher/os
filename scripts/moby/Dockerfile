FROM rancher/os

# replace this with `rancher/os-initrd`
RUN mkdir /tmp/initrd \
  && cd /tmp/initrd \
  && cat /dist/initrd-* | gunzip | cpio -i \
  && rm -rf usr/lib \
  && mv /sbin /tmp/sbin_bak \
  && mv /usr/sbin /tmp/usr_sbin_bak \
  && rm /tmp/initrd/usr/var/lib/cni/bin/host-local /tmp/initrd/usr/var/lib/cni/bin/bridge \
  && mkdir -p /tmp/initrd/var/lib/cni/bin \
  && ln -s ../../../../usr/bin/ros /tmp/initrd/var/lib/cni/bin/host-local \
  && ln -s ../../../../usr/bin/ros /tmp/initrd/var/lib/cni/bin/bridge \
  && rm /tmp/initrd/usr/share/ros/images.tar \
  && cp -r --update --force /tmp/initrd/* / \
  && cp -r --update --force /tmp/sbin_bak/* /sbin/ \
  && cp -r --update --force /tmp/usr_sbin_bak/* /usr/sbin/ \
  && rm /usr/bin/depmod /usr/bin/insmod /usr/bin/lsmod /usr/bin/modinfo /usr/bin/modprobe /usr/bin/rmmod \
  && ln -s /usr/bin/busybox /usr/bin/depmod \
  && ln -s /usr/bin/busybox /usr/bin/insmod \
  && ln -s /usr/bin/busybox /usr/bin/lsmod \
  && ln -s /usr/bin/busybox /usr/bin/modinfo \
  && ln -s /usr/bin/busybox /usr/bin/modprobe \
  && ln -s /usr/bin/busybox /usr/bin/rmmod \
  && rm -rf /dist /scripts /tmp/* \
  && cd / \
  && rm -rf /tmp/initrd \
  && mkdir -p /lib/bootchart/tmpfs


#FROM rancher/os-installer
#RUN cp /bin/ros /init

ADD bootchart-collector /lib/bootchart/
ADD bootchartd /sbin/
ADD bootchartd.conf /etc/
RUN rm -f /init /sbin/init \
    && ln -s /usr/bin/ros /sbin/init
ADD bootchartd /init
