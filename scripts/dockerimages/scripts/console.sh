#!/bin/sh

#if [ -t 1 ]; then
    #exec /bin/sh
#else
    exec respawn << EOF
/sbin/getty 115200 tty1
/sbin/getty 115200 tty2
/sbin/getty 115200 tty3
/sbin/getty 115200 tty4
/sbin/getty 115200 tty5
/sbin/getty 115200 tty6
EOF
#fi
