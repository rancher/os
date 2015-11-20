# This file describes the standard way to build libcompose, using docker
FROM golang:1.4.2-cross

RUN apt-get update && apt-get install -y \
    iptables \
    build-essential \
    --no-install-recommends

# Install build dependencies
RUN go get github.com/mitchellh/gox
RUN go get github.com/aktau/github-release
RUN go get github.com/tools/godep
RUN go get golang.org/x/tools/cmd/cover
RUN go get github.com/golang/lint/golint
RUN go get golang.org/x/tools/cmd/vet

# Which docker version to test on
ENV DOCKER_VERSION 1.7.1

# Download docker
RUN set -ex; \
    curl https://get.docker.com/builds/Linux/x86_64/docker-${DOCKER_VERSION} -o /usr/local/bin/docker-${DOCKER_VERSION}; \
    chmod +x /usr/local/bin/docker-${DOCKER_VERSION}

# Set the default Docker to be run
RUN ln -s /usr/local/bin/docker-${DOCKER_VERSION} /usr/local/bin/docker

ENV GOPATH /go/src/github.com/docker/libcompose/Godeps/_workspace:/go
ENV COMPOSE_BINARY /go/src/github.com/docker/libcompose/libcompose-cli
ENV USER root

WORKDIR /go/src/github.com/docker/libcompose

# Wrap all commands in the "docker-in-docker" script to allow nested containers
ENTRYPOINT ["script/dind"]

COPY . /go/src/github.com/docker/libcompose
