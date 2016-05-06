#!/bin/bash
set -x

if [ "$(ros config get rancher.state.mdadm_scan)" = "true" ]; then
    mdadm --assemble --scan
fi

ros config get rancher.state.script > config.sh
if [ -s config.sh ]; then
    chmod +x config.sh
    exec ./config.sh
fi
