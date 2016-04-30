#!/bin/bash
set -x -e

dhcp -daemon=${DAEMON:-false}
