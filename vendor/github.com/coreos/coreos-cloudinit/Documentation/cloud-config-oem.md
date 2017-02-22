## OEM configuration

The `coreos.oem.*` parameters follow the [os-release spec][os-release], but have been repurposed as a way for coreos-cloudinit to know about the OEM partition on this machine. Customizing this section is only needed when generating a new OEM of CoreOS from the SDK. The fields include:

- **id**: Lowercase string identifying the OEM
- **name**: Human-friendly string representing the OEM
- **version-id**: Lowercase string identifying the version of the OEM
- **home-url**: Link to the homepage of the provider or OEM
- **bug-report-url**: Link to a place to file bug reports about this OEM

coreos-cloudinit renders these fields to `/etc/oem-release`.
If no **id** field is provided, coreos-cloudinit will ignore this section.

For example, the following cloud-config document...

```yaml
#cloud-config
coreos:
  oem:
    id: "rackspace"
    name: "Rackspace Cloud Servers"
    version-id: "168.0.0"
    home-url: "https://www.rackspace.com/cloud/servers/"
    bug-report-url: "https://github.com/coreos/coreos-overlay"
```

...would be rendered to the following `/etc/oem-release`:

```yaml
ID=rackspace
NAME="Rackspace Cloud Servers"
VERSION_ID=168.0.0
HOME_URL="https://www.rackspace.com/cloud/servers/"
BUG_REPORT_URL="https://github.com/coreos/coreos-overlay"
```

[os-release]: http://www.freedesktop.org/software/systemd/man/os-release.html
