FROM rancher/os-base
COPY syslog.sh /
RUN sed -i 1,10d /etc/rsyslog.conf
CMD ["/syslog.sh"] 
