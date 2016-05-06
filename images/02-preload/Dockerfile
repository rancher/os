FROM rancher/os-base
RUN ln -sf /usr/bin/docker.dist /usr/bin/docker
COPY preload.sh  /
CMD ["/preload.sh"]
