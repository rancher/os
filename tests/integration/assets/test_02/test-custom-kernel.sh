#!/bin/bash
set -ex

cd $(dirname $0)/../../../..

cp ./tests/integration/assets/test_02/build.conf ./

make -f Makefile.docker minimal

exec ./scripts/run --qemu --no-rebuild --fresh
