FROM rancher/os-base
ENV KERNEL_HEADERS_URL https://github.com/rancher/os-kernel/releases/download/Ubuntu-4.4.0-47.68-rancher1/build.tar.gz
RUN wget -O /build.tar.gz $KERNEL_HEADERS_URL
COPY headers.sh  /
CMD ["/headers.sh"]
