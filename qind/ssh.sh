#!/bin/sh
set -e

exec ssh -F /ssh_config "$@"
