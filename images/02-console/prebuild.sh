#!/bin/bash
set -e

VERSION=${VERSION:?"VERSION not set"}

cd $(dirname $0)

rm -rf ./build
mkdir -p ./build

cat > ./build/lsb-release << EOF
DISTRIB_ID=${DISTRIB_ID}
DISTRIB_RELEASE=${VERSION}
DISTRIB_DESCRIPTION="${DISTRIB_ID} ${VERSION}"
EOF

cat > ./build/sshd_config.append.tpl << EOF
{{- if .Port}}
Port {{.Port}}
{{- end}}

{{- if .ListenAddress}}
ListenAddress {{.ListenAddress}}
{{- end}}

ClientAliveInterval 180

UseDNS no
PermitRootLogin no
AllowGroups docker
EOF
