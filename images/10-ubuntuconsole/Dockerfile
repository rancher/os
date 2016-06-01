FROM rancher/os-ubuntuconsole-base
# FROM amd64=ubuntu:16.04 arm64=aarch64/ubuntu:16.04 arm=armhf/ubuntu:16.04
RUN apt-get update && \
    apt-get upgrade --no-install-recommends -y && \
    apt-get install -y --no-install-recommends iptables openssh-server rsync vim curl ca-certificates psmisc htop
RUN rm -rf /etc/ssh/*key*
RUN rm -fr /sbin/poweroff  /sbin/shutdown /sbin/reboot /sbin/halt /usr/sbin/poweroff  /usr/sbin/shutdown /usr/sbin/reboot /usr/sbin/halt
COPY build/entry.sh build/console.sh build/docker-init build/update-ssh-keys build/rancheros-install /usr/sbin/
RUN locale-gen en_US.UTF-8
RUN addgroup --gid 1100 rancher && \
    addgroup --gid 1101 docker && \
    useradd -u 1100 -g rancher -G docker,sudo -m -s /bin/bash rancher && \
    useradd -u 1101 -g docker -G docker,sudo -m -s /bin/bash docker && \
    echo ClientAliveInterval 180 >> /etc/ssh/sshd_config && \
    echo '## allow password less for rancher user' >> /etc/sudoers && \
    echo 'rancher ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers && \
    echo '## allow password less for docker user' >> /etc/sudoers && \
    echo 'docker ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers && \
    ln -sf /usr/bin/docker.dist /usr/bin/docker && \
    ln -sf /usr/bin/docker-containerd.dist /usr/bin/docker-containerd && \
    ln -sf /usr/bin/docker-containerd-shim.dist /usr/bin/docker-containerd-shim && \
    ln -sf /usr/bin/docker-runc.dist /usr/bin/docker-runc
ENTRYPOINT ["/usr/sbin/entry.sh"]
CMD ["/usr/sbin/console.sh"]
