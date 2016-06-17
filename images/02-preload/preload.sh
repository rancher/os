#!/bin/bash
set -e

BASE=${1:-${PRELOAD_DIR}}
BASE=${BASE:-/mnt/preload}

should_load() {
    file=${1}
    if [[ ${file} =~ \.done$ ]]; then echo false
    elif [ -f ${file} ]; then
        if [[ ${file} -nt ${file}.done ]]; then echo true
        else echo false
        fi
    else echo false
    fi
}

if [ -d ${BASE} ]; then
    echo Preloading docker images from ${BASE}...

    for file in $(ls ${BASE}); do
        path=${BASE}/${file}
        loading=$(should_load ${path})
        if [ ${loading} == "true" ]; then
            CAT="cat ${path}"
            if [[ ${file} =~ \.t?gz$ ]]; then CAT="${CAT} | gunzip"; fi
            if [[ ${file} =~ \.t?xz$ ]]; then CAT="${CAT} | unxz"; fi
            wait-for-docker
            CAT="${CAT} | docker load"
            echo loading from ${path}
            eval ${CAT} || :
            touch ${path}.done || :
        fi
    done

    echo Done.
else
    echo Can not preload images from ${BASE}: not a dir or does not exist.
fi

