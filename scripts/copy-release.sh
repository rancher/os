#!/bin/bash
. ./scripts/version
gsutil -m cp -r dist/artifacts/* gs://releases.rancher.com/os/${VERSION}
