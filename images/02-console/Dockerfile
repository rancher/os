FROM rancher/os-base
COPY build/lsb-release /etc/
RUN sed -i 's/rancher:!/rancher:*/g' /etc/shadow && \
    sed -i 's/docker:!/docker:*/g' /etc/shadow && \
    sed -i 's/#ClientAliveInterval 0/ClientAliveInterval 180/g' /etc/ssh/sshd_config && \
    echo '## allow password less for rancher user' >> /etc/sudoers && \
    echo 'rancher ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers && \
    echo '## allow password less for docker user' >> /etc/sudoers && \
    echo 'docker ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers
COPY prompt.sh /etc/profile.d/
