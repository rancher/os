# Dashboard/UI

The Rancher UI is running by default on port `:8443`.  There is no default
`admin` user password set.  You must run `rancherd reset-admin` once to
get an `admin` password to login.

To disable the Rancher UI from running on a host port, or to change the
default hostPort used the below configuration.

```yaml
#cloud-config
rancherd:
  rancherValues:
    # Setting the host port to 0 will disable the hostPort
    hostPort: 0
```
