# Supported Versions and Channels

The `kubernetesVersion` and `rancherVersion` fields accept explicit versions
numbers or channel names.

## Valid Versions

The list of valid versions for the `kubernetesVersion` field can be determined
from the Rancher metadata using the following commands.

__k3s:__
```bash
curl -sL https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/release-v2.6/data/data.json | jq -r '.k3s.releases[].version'
```
__rke2:__
```bash
curl -sL https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/release-v2.6/data/data.json | jq -r '.rke2.releases[].version'
```

The list of valid `rancherVersion` values can be obtained from the
[stable](https://artifacthub.io/packages/helm/rancher-stable/rancher) and
[latest](https://artifacthub.io/packages/helm/rancher-latest/rancher) helm
repos. The version string is expected to be the "application version" which
is the version starting with a `v`. For example, `v2.6.2` is the current
format not `2.6.2`.

## Version Channels 

Valid `kubernetesVersion` channels are as follows:

| Channel Name | Description |
|--------------|-------------|
|  stable | k3s stable (default value of kubernetesVersion) |
| latest | k3s latest |
| testing | k3s test |
|  stable:k3s | Same as stable channel |
| latest:k3s | Same as latest channel |
| testing:k3s | Same as testing channel |
|  stable:rke2 | rke2 stable |
| latest:rke2 | rke2 latest |
| testing:rke2 | rke2 testing |
| v1.21 | Latest k3s v1.21 release. The applies to any Kubernetes minor version |
| v1.21:rke2 | Latest rke2 v1.21 release. The applies to any Kubernetes minor version |

Valid `rancherVersions` channels are as follows:

| Channel Name | Description |
|--------------|-------------|
|  stable | [stable helm repo](https://artifacthub.io/packages/helm/rancher-stable/rancher) |
| latest | [latest helm repo](https://artifacthub.io/packages/helm/rancher-latest/rancher) |

