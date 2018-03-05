RaspberryPi 3 Image (ARMv8 | AARCH64 | ARM64)
=============================================

Build by running `dapper` in this folder and the build will produce `./dist/rancheros-raspberry-pi64.zip`.

This image is compatible with the Raspberry Pi 3, since it is the only ARMv8 device in the Raspberry Pi family at the moment.

Build Requirements
==================

This build uses local loopback devices and thus requires to run as a privileged container.  So please keep the setting `ENV DAPPER_RUN_ARGS --privileged` from `Dockerfile.dapper` for now.  The build is running quite fast and has been tested on OS X with boot2docker.
