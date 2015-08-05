#!/bin/bash

export NO_TEST=true
exec $(dirname $0)/scripts/ci
