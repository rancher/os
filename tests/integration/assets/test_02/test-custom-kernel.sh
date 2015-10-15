#!/bin/bash
set -ex

cd $(dirname $0)/../../../..

cp ./tests/integration/assets/test_02/build.conf ./

make -f Makefile.docker DEV_BUILD=1 minimal

exec ./scripts/run --qemu --no-rebuild --no-rm-usr --fresh
