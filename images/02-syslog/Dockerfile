FROM rancher/os-base
COPY logrotate.d/ /usr/share/logrotate/logrotate.d/
COPY entrypoint.sh /usr/bin/entrypoint.sh
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
