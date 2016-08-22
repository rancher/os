FROM rancher/os-base
RUN ln -sf /var/lib/rancher/engine/docker /usr/bin/docker
COPY preload.sh  /
CMD ["/preload.sh"]
