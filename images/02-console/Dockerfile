FROM debian:stable-slim

COPY build/sshd_config.append.tpl /etc/ssh/
COPY build/lsb-release /etc/

RUN apt-get update \
    && apt-get install -y --no-install-recommends iptables openssh-server rsync locales \
                           sudo less curl ca-certificates psmisc htop kmod iproute2 \
                           net-tools bash-completion wget \
                           nano open-iscsi iputils-ping \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /etc/ssh/*key* \
    && echo 'en_US.UTF-8 UTF-8' > /etc/locale.gen \
    && locale-gen \
    && find /usr/share/i18n/charmaps -not -path /usr/share/i18n/charmaps/UTF-8.gz -name '*.gz' -exec rm -rf {} \; \
    && find /usr/share/i18n/locales -not -path /usr/share/i18n/locales/en_US -name '*_*' -exec rm -rf {} \; \
    && addgroup --gid 1100 rancher \
    && addgroup --gid 1101 docker \
    && useradd -u 1100 -g rancher -G docker,sudo -m -s /bin/bash rancher \
    && useradd -u 1101 -g docker -G docker,sudo -m -s /bin/bash docker \
    && echo '## allow password less for rancher user' >> /etc/sudoers \
    && echo 'rancher ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers \
    && echo '## allow password less for docker user' >> /etc/sudoers \
    && echo 'docker ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers \
    && ln -s /bin/ps /usr/bin/ps \
    && cat /etc/ssh/sshd_config > /etc/ssh/sshd_config.tpl \
    && cat /etc/ssh/sshd_config.append.tpl >> /etc/ssh/sshd_config.tpl \
    && rm -f /etc/ssh/sshd_config.append.tpl /etc/ssh/sshd_config \
    && rm -f /etc/motd \
    && ln -s /etc/issue /etc/motd
ENTRYPOINT ["/usr/bin/ros", "entrypoint"]
