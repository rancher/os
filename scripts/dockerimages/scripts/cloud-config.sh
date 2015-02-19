#!/bin/bash

set -x -e

CLOUD_CONFIG_FLAGS=$(rancherctl config get cloud_config)

cloudinit --preinit "$CLOUD_CONFIG_FLAGS"
