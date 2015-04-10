#!/bin/bash

if [ -e /host/dev ]; then
    mount --rbind /host/dev /dev
fi

CA_BASE=/etc/ssl/certs/ca-certificates.crt.rancher
CA=/etc/ssl/certs/ca-certificates.crt

if [[ -e ${CA_BASE} && ! -e ${CA} ]]; then
    cp $CA_BASE $CA
fi

exec "$@"
