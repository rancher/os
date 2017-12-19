#!/bin/bash

cd $(dirname $0)/..

cat hack/vendor.sh | grep '^clone git' | sed 's/#.*//' | awk '{print $3 " " $4 " " $5}' | sort > trash.conf

rm -rf \
    api/client/trust.go \
    api/server/router/network \
    contrib \
    daemon/container_operations.go \
    daemon/discovery.go \
    daemon/discovery_test.go \
    daemon/network.go \
    daemon/graphdriver/aufs \
    daemon/graphdriver/btrfs \
    daemon/graphdriver/devmapper \
    daemon/graphdriver/windows \
    daemon/graphdriver/zfs \
    daemon/graphdriver/register/register_aufs.go \
    daemon/graphdriver/register/register_btrfs.go \
    daemon/graphdriver/register/register_devicemapper.go \
    daemon/graphdriver/register/register_windows.go \
    daemon/graphdriver/register/register_zfs.go \
    daemon/logger/gelf \
    daemon/logger/awslogs \
    daemon/logger/etwlogs \
    daemon/logger/fluentd \
    daemon/logger/gcplogs \
    daemon/logger/journald \
    daemon/logger/splunk \
    docs \
    experimental \
    hack \
    integration-cli \
    man \
    migrate \
    pkg/discovery \
    pkg/term/windows \
    project \
    vendor \
    Dockerfile* \
    AUTHORS \
    CHANGELOG.md \
    CONTRIBUTING.md \
    MAINTAINERS \
    README.md \
    ROADMAP.md \
    VERSION \
    VENDORING.md \
    .mailmap

find -name '*_windows.go' -exec rm -f {} \;

sed -i 's/^package main/package docker/g' docker/*.go

find -name "*.go" -exec grep os/exec {} \; -exec sed -i 's!os/exec!github.com/docker/containerd/subreaper/exec!g' {} \;
