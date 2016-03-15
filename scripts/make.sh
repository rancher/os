#!/bin/bash
set -e

cd $(dirname $0)/..
. ./scripts/dapper-common

dapper make HOST_ARCH=${HOST_ARCH} ARCH=${ARCH} "$@"
