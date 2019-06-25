FROM rancher/os-base
COPY . /
RUN chmod 644 /etc/logrotate.conf
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
