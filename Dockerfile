FROM debian:jessie
COPY ./scripts/installer /scripts
COPY ./scripts/version /scripts/
RUN /scripts/bootstrap

COPY ./dist/artifacts/vmlinuz /dist/vmlinuz
COPY ./dist/artifacts/initrd  /dist/initrd

ENTRYPOINT ["/scripts/lay-down-os"]
