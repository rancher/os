# VMWare Guestinfo Interface

## Cloud-Config VMWare Guestinfo Variables

coreos-cloudinit accepts configuration from the VMware RPC API's *guestinfo*
facility. This datasource can be enabled with the `--from-vmware-guestinfo`
flag to coreos-cloudinit.

The following guestinfo variables are recognized and processed by cloudinit
when passed from the hypervisor to the virtual machine at boot time. Note that
property names are prefixed with `guestinfo.` in the VMX, e.g., `guestinfo.hostname`.

|            guestinfo variable             |              type               |
|:--------------------------------------|:--------------------------------|
| `hostname`                            | `hostname`                      |
| `interface.<n>.name`                  | `string`                        |
| `interface.<n>.mac`                   | `MAC address`                   |
| `interface.<n>.dhcp`                  | `{"yes", "no"}`                 |
| `interface.<n>.role`                  | `{"public", "private"}`         |
| `interface.<n>.ip.<m>.address`        | `CIDR IP address`               |
| `interface.<n>.route.<l>.gateway`     | `IP address`                    |
| `interface.<n>.route.<l>.destination` | `CIDR IP address`               |
| `dns.server.<x>`                      | `IP address`                    |
| `dns.domain.<y>`                      | `DNS search domain`             |
| `coreos.config.data`                  | `string`                        |
| `coreos.config.data.encoding`         | `{"", "base64", "gzip+base64"}` |
| `coreos.config.url`                   | `URL`                           |

Note: "n", "m", "l", "x" and "y" are 0-indexed, incrementing integers. The
identifier for an `interface` does not correspond to anything outside of this
configuration; it serves only to distinguish between multiple `interface`s.

The guide to [booting on VMWare][bootvmware] is the starting point for more
information about configuring and running CoreOS on VMWare.

[bootvmware]: https://github.com/coreos/docs/blob/master/os/booting-on-vmware.md
