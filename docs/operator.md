# RancherOS Operator

The RancherOS operator is responsible for managing the RancherOS versions
and maintaining a machine inventory to assist with secure device on-boarding.

## Managing Upgrades

The RancherOS will manage the upgrade of the local cluster where the operator
is running and also any downstream cluster managed by Rancher Multi-Cluster
Manager.

### ManagedOSImage

The ManagedOSImage kind used to define what version of RancherOS should be
running on each node. The simplest example of this type would be to change
the version of the local nodes.

```bash
kubectl edit -n fleet-local defautl-os-image
```
```yaml
apiVersion: rancheros.cattle.io/v1
kind: ManagedOSImage
metadata:
  name: default-os-image
  namespace: fleet-local
spec:
  osImage: rancher/os2:v0.0.0
```


#### Reference

Below is reference of the full type

```yaml
apiVersion: rancheros.cattle.io/v1
kind: ManagedOSImage
metadata:
  name: arbitrary
  
  # There are two special namespaces to consider.  If you wish to manage
  # nodes on the local cluster this namespace should be `fleet-local`. If
  # you wish to manage nodes in Rancher MCM managed clusters then the
  # namespace is typically fleet-default.
  namespace: fleet-local
spec:
  # The image name to pull for the OS
  osImage: rancher/os2:v0.0.0
  
  # The selector for which nodes will be select.  If null then all nodes
  # will be selected
  nodeSelector:
    matchLabels: {}
    
  # How many nodes in parallel to update.  If empty the default is 1 and
  # if set to 0 the rollout will be paused
  concurrency: 2
    
  # Arbitrary action to perform on the node prior to upgrade
  prepare:
    image: ...
    command: ["/bin/sh"]
    args: ["-c", "true"]
    env:
    - name: TEST_ENV
      value: testValue
      
  # Parameters to control the drain behavior.  If null no draining will happen
  # on the node.
  drain:
    # Refer to kubectl drain --help for the definition of these values
    timeout: 5m
    gracePeriod: 5m
    deleteLocalData: false
    ignoreDaemonSets: true
    force: false
    disableEviction: false
    skipWaitForDeleteTimeout: 5
    
  # Which cluster to target
  # This is used if you are running Rancher MCM and managing
  # multiple clusters.  The syntax of this field matches the
  # Fleet targets and is described at https://fleet.rancher.io/gitrepo-targets/
  targets: []
```