FROM rancher/os-base
RUN ln -sf /usr/bin/docker.dist /usr/bin/docker && \
    ln -sf /usr/bin/docker-containerd.dist /usr/bin/docker-containerd && \
    ln -sf /usr/bin/docker-containerd-shim.dist /usr/bin/docker-containerd-shim && \
    ln -sf /usr/bin/docker-runc.dist /usr/bin/docker-runc
COPY preload.sh  /
CMD ["/preload.sh"]
