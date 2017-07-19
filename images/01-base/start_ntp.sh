#!/bin/sh
set -ex
echo "starting in one shot mode to fix large time differences"
ntpd -gq
echo "starting long running nptd"
exec ntpd --nofork -g
