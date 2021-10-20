# RancherOS v2

RancherOS v2 is an immutable Linux distribution built to run Rancher and
it's corresponding Kubernetes distributions [RKE2](https://rke2.io) 
and [k3s](https://k3s.io). It is built using the [cOS-toolkit](https://rancher-sandbox.github.io/cos-toolkit-docs/docs/)
and based on openSUSE. Initial node configurations is done using only a
cloud-init style approach and all further maintenance is done using
Kubernetes operators.

## Use Cases

RancherOS is intended to be ran as the operating system beneath a Rancher Multi-Cluster 
Management server or as a node in a Kubernetes cluster managed by Rancher. RancherOS
also allows you to build stand alone Kubernetes clusters that run an embedded
and smaller version of Rancher to manage the local cluster. A key attribute of RancherOS
is that it is managed by Rancher and thus Rancher will exist either locally in the cluster
or centrally with Rancher Multi-Cluster Manager.

## Architecture

### OCI Image based

RancherOS v2 is an A/B style image based distribution. One first runs
on a read-only image A and to do an upgrade pulls a new read only image
B and then reboots the system to run on B. What is unique about
RancherOS v2 is that the runtime images come from OCI Images. Not an
OCI Image containing special artifacts, but an actual Docker runnable
image that is built using standard Docker build processes. RancherOS is
built using normal `docker build` and if you wish to customize the OS
image all you need to do is create a new `Dockerfile`.

### rancherd

RancherOS v2 includes no container runtime, Kubernetes distribution,
or Rancher itself. All of these assests are dynamically pulled at runtime. All that
is included in RancherOS is [rancherd](https://github.com/rancher/rancherd) which
is responsible for bootstrapping RKE2/k3s and Rancher from an OCI registry. This means
an update to containerd, k3s, RKE2, or Rancher does not require an OS upgrade
or node reboot.

### cloud-init

RancherOS v2 is initially configured using a simple version of `cloud-init`.
It is not expected that one will need to do a lot of customization to RancherOS
as the core OS's sole purpose is to run Rancher and Kubernetes and not serve as
a generic Linux distribution.

### RancherOS Operator

RancherOS v2 includes an operator that is responsible for managing OS upgrades
and assiting with secure device onboarding (SDO).


### openSUSE Leap

RancherOS v2 is based off of openSUSE Leap.  There is no specific tie in to
openSUSE beyond that RancherOS assumes the underlying distribution is
based on systemd. We choose openSUSE for obvious reasons, but beyond
that openSUSE Leap provides a stable layer to build upon that is well
tested and has paths to commercial support, if one chooses.