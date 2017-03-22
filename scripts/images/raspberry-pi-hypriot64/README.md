RaspberryPi 2 Image
===================

Build by running `dapper` in this folder and the build will produce `./dist/rancheros-rpi2.zip`.

This image is compatible with the Raspberry Pi 3 too, but only ARMv7 is supported now.

Build Requirements
==================

This build uses local loopback devices and thus requires to run as a privileged container.  So please keep the setting `ENV DAPPER_RUN_ARGS --privileged` from `Dockerfile.dapper` for now.  The build is running quite fast and has been tested on OS X with boot2docker.
