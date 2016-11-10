FROM alpine
RUN apk --no-cache add qemu-system-x86_64 qemu-img cdrkit openssh-client bash
COPY . /
#COPY . /
RUN mkdir -p /assets /dist/artifacts /build /bin /scripts
ENTRYPOINT ["/scripts/run"]
