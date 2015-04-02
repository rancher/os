#!/bin/sh

# udhcpc script edited by Tim Riker <Tim@Rikers.org>

[ -z "$1" ] && echo "Error: should be called from udhcpc" && exit 1

RESOLV_CONF="/etc/resolv.conf"
[ -e $RESOLV_CONF ] || touch $RESOLV_CONF
[ -n "$broadcast" ] && BROADCAST="broadcast $broadcast"
[ -n "$subnet" ] && NETMASK="netmask $subnet"


case "$1" in
        deconfig)
                /sbin/ifconfig $interface up
                /sbin/ifconfig $interface 0.0.0.0

                # drop info from this interface
                # resolv.conf may be a symlink to /tmp/, so take care
                TMPFILE=$(mktemp)
                grep -vE "# $interface\$" $RESOLV_CONF > $TMPFILE
                cat $TMPFILE > $RESOLV_CONF
                rm -f $TMPFILE

                if [ -x /usr/sbin/avahi-autoipd ]; then
                        /usr/sbin/avahi-autoipd -k $interface
                fi
                ;;

        leasefail|nak)
                if [ -x /usr/sbin/avahi-autoipd ]; then
                        /usr/sbin/avahi-autoipd -wD $interface --no-chroot
                fi
                ;;

        renew|bound)
                if [ -x /usr/sbin/avahi-autoipd ]; then
                        /usr/sbin/avahi-autoipd -k $interface
                fi
                /sbin/ifconfig $interface $ip $BROADCAST $NETMASK
                /sbin/ifconfig $interface mtu $mtu

                if [ -n "$router" ] ; then
                        echo "deleting routers"
                        while route del default gw 0.0.0.0 dev $interface 2> /dev/null; do
                                :
                        done

                        for i in $router ; do
                                ip route add $i/32 dev $interface
                                ip route add default via $i
                        done
                fi

                # drop info from this interface
                # resolv.conf may be a symlink to /tmp/, so take care
                TMPFILE=$(mktemp)
                grep -vE "# $interface\$" $RESOLV_CONF > $TMPFILE
                cat $TMPFILE > $RESOLV_CONF
                rm -f $TMPFILE

                [ -n "$domain" ] && echo "search $domain # $interface" >> $RESOLV_CONF
                for i in $dns ; do
                        echo adding dns $i
                        echo "nameserver $i # $interface" >> $RESOLV_CONF
                done
                ;;
esac

HOOK_DIR="$0.d"
for hook in "${HOOK_DIR}/"*; do
    [ -f "${hook}" -a -x "${hook}" ] || continue
    "${hook}" "${@}"
done

exit 0
