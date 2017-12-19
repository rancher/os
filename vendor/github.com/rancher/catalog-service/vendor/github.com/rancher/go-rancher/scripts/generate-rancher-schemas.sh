#!/bin/bash
set -e -x

cd $(dirname $0)/../generator

URL_BASE='http://localhost:8080'

if [ "$1" != "" ]; then
    URL_BASE=$1
fi

echo -n Waiting for cattle ${URL_BASE}/ping
while ! curl -fs ${URL_BASE}/ping; do
    echo -n .
    sleep 1
done
echo

source $(dirname "$0")/../scripts/common_functions

gen ${URL_BASE}/v1-catalog catalog rename
gen ${URL_BASE}/v2-beta

echo Success
