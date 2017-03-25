#!/bin/bash

#
# This script will make a shell script that can be used as a cloud-init style user data script
# or run as root from a debian/ubuntu DigitalOcean VM to replace that distribution with
# RancherOS
#
# Its intended to be used for development, but can easily be modified to be more generally
# useful - make a Pull Request :)
# 
# Note: this script will run caddy in your os/dist/artifacts/ directory, so don't leave it
# running unsupervised.

DIST="../../../dist/artifacts"
command -v caddy >/dev/null 2>&1 || { echo >&2 "I require caddy but it's not installed, see https://github.com/mholt/caddy#quick-start . Aborting."; exit 1; }

if [[ ! -e "$DIST" ]]; then
	echo "Need to 'make release' so that there are files to serve. Aborting."
	exit 1
fi

source ${DIST}/../../scripts/version
VMLINUX=$(ls -1 ${DIST}/vmlinuz-* | head -n1)
INITRD="${DIST}/initrd-${VERSION}"

IP=$(curl ipinfo.io/ip)
PORT=2115

CLOUDCONFIG="digitalocean.sh"

cat cloud-config.yml \
	| sed "s|^URL_BASE.*$|URL_BASE=http://${IP}:${PORT}|g" \
	| sed "s|^VMLINUX.*$|VMLINUX=${VMLINUX}|g" \
	| sed "s|^INITRD.*$|INITRD=${INITRD}|g" \
		> ${DIST}/${CLOUDCONFIG}

echo "Hosting a cloud-config script at http://${IP}:${PORT}/${CLOUDCONFIG}"

cd ${DIST}
caddy -port ${PORT}
