RaspberryPi 2 Image
===================

Build by running `dapper` in this folder and the build will produce `./dist/rancheros-rpi2.zip`.

KVM
===

This build requires a host capable of KVM.  If you don't have KVM then remove `ENV DAPPER_RUN_ARGS --device /dev/kvm` from `Dockerfile.dapper`, but it will run very slow.
