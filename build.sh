#!/bin/bash
set -e

if [ "$1" != "--dev" ]; then
  echo
  echo Running \"production\" build. Will use lzma to compress initrd, which is somewhat slow...
  echo Ctrl+C if you don\'t want this.
  echo
  echo For \"developer\" builds, run ./build.sh --dev
  echo
  dapper -k make all
else
  dapper -k make DEV_BUILD=1 all
fi


ls -lh dist/artifacts
