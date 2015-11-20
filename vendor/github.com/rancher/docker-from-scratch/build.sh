#!/bin/bash

rm -rf $(dirname $0)/build

export NO_TEST=true
exec $(dirname $0)/scripts/ci
