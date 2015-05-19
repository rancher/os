#!/bin/bash
set -e

BASE=${1:-${PRELOAD_DIR}}
BASE=${BASE:-/mnt/preload}

if [ -d ${BASE} ]; then
    echo Preloading docker images from ${BASE}...

    for file in $(ls ${BASE}); do
        if [ -f ${BASE}/${file} ]; then
            CAT="cat ${BASE}/${file}"
            if [[ ${file} =~ \.t?gz$ ]]; then CAT="${CAT} | gunzip"; fi
            if [[ ${file} =~ \.t?xz$ ]]; then CAT="${CAT} | unxz"; fi
            CAT="${CAT} | docker load"
            echo loading from ${BASE}/${file}
            eval ${CAT} || :
        fi
    done

    echo Done.
else
    echo Can not preload images from ${BASE}: not a dir or does not exist.
fi

