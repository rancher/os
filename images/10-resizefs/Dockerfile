FROM rancher/os-resizefs-base
# FROM amd64=ubuntu:16.04 arm64=aarch64/ubuntu:16.04 arm=armhf/ubuntu:16.04
RUN apt-get update && \
    apt-get install -y --no-install-recommends cloud-guest-utils parted && \
    apt-get clean && rm -rf /var/lib/apt/lists/*
COPY resizefs.sh /
ENTRYPOINT ["/resizefs.sh"]
