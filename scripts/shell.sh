#!/bin/bash
set -e

cd $(dirname $0)/..
. ./scripts/dapper-common

exec dapper -d -s
