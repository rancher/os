#!/bin/bash
set -x -e

CGROUPS="perf_event net_cls freezer devices blkio memory cpuacct cpu cpuset"

mkdir -p /sys/fs/cgroup
mount -t tmpfs none /sys/fs/cgroup

for i in $CGROUPS; do
    mkdir -p /sys/fs/cgroup/$i
    mount -t cgroup -o $i none /sys/fs/cgroup/$i
done

if ! lsmod | grep -q br_netfilter; then
    modprobe br_netfilter 2>/dev/null || true
fi

rm -f /var/run/docker.pid

USE_TLS=$(rancherctl config get userdocker.use_tls)

if [ "$USE_TLS" == "true" ]; then
    TLS_CA_CERT=$(rancherctl config get userdocker.tls_ca_cert)
    TLS_SERVER_CERT=$(rancherctl config get userdocker.tls_server_cert)
    TLS_SERVER_KEY=$(rancherctl config get userdocker.tls_server_key)    

    TLS_PATH=/etc/docker/tls
    mkdir -p $TLS_PATH 

    if [ -n "$TLS_CA_CERT" ] && [ -n "$TLS_SERVER_CERT" ] && [ -n "$TLS_SERVER_KEY" ]; then
	echo "$TLS_CA_CERT" > $TLS_PATH/ca.pem
	echo "$TLS_SERVER_CERT" > $TLS_PATH/server-cert.pem
	echo "$TLS_SERVER_KEY" > $TLS_PATH/server-key.pem
    else
        tlsconf
    	TLS_CA_CERT="$(cat $TLS_PATH/ca.pem)"
    	TLS_SERVER_CERT="$(cat $TLS_PATH/server-cert.pem)"
    	TLS_SERVER_KEY="$(cat $TLS_PATH/server-key.pem)"
    fi 
    
    rancherctl config set -- userdocker.tls_ca_cert "$TLS_CA_CERT"
    rancherctl config set -- userdocker.tls_server_cert "$TLS_SERVER_CERT"
    rancherctl config set -- userdocker.tls_server_key "$TLS_SERVER_KEY"

    exec docker -d -s overlay --tlsverify --tlscacert=$TLS_PATH/ca.pem --tlscert=$TLS_PATH/server-cert.pem --tlskey=$TLS_PATH/server-key.pem -H=0.0.0.0:2376 -H=unix:///var/run/docker.sock
else
    exec docker -d -s overlay
fi
