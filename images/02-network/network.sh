#!/bin/bash
set -x -e

netconf -daemon=${DAEMON:-false}
