#!/bin/bash
set -e

cd $(dirname $0)/..
. ./scripts/dapper-common

dapper make ARCH=${ARCH} "$@"
