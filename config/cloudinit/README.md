# cloudinit

Derived from [coreos-cloudinit](https://github.com/coreos/coreos-cloudinit) (see LICENSE and NOTICE).
BurmillaOS only keeps the parts `ros` uses: the metadata datasources (`datasource/`),
the `#cloud-config` / `#!` user-data detection, the `write_files` type and the
`$public_ipv4` / `$private_ipv4` substitution (`initialize/`). The CoreOS unit, etcd,
fleet, flannel and user handling was removed.

The cloud-config options BurmillaOS supports are documented at
https://burmillaos.org/docs/configuration/.
