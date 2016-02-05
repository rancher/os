#!/bin/bash

cd $(dirname $0)
rm -rf ./build

export NO_TEST=true
dapper ./scripts/ci
