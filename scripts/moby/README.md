# RancherOS built using LinuxKit/Moby

This is an initial non-containerd version.

To build, first run `make dev`, and then `make build-moby`.

To run in qemu, use `make run-moby`.

At the moment, `linuxkit run rancheros` crashes with a kernel panic.
