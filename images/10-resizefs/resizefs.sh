#!/bin/bash
set -ex

RESIZE_DEV=${RESIZE_DEV:?"RESIZE_DEV not set."}
STAMP=/var/log/resizefs.done

if [ -e "${STAMP}" ]; then
    echo FS already resized.
    exit 0
fi

# TODO: rm hardcoded partition number, maybe identify RANCHER_STATE partition (can be the whole device)
if [ -b "${RESIZE_DEV}" ]; then
  growpart ${RESIZE_DEV} 1 || :  # ignore error "NOCHANGE: partition 1 is size NNN. it cannot be grown"
  partprobe ${RESIZE_DEV}
  resize2fs ${RESIZE_DEV}1
else
  echo "Block device expected: ${RESIZE_DEV} is not."
  exit 1
fi

touch $STAMP
