# Understanding Clusters

Rancherd bootstraps a node with Kubernetes (k3s/rke2) and Rancher such
that all future management of Kubernetes and Rancher can be done from
Kubernetes. Rancherd will only run once per node. Once the system has
been fully bootstrapped it will not run again. It is intended that the
primary use of Rancherd is to be ran from cloud-init or a similar system.

## Cluster Initialization

Creating a cluster always starts with one node initializing the cluster, by
assigning the `cluster-init` role and then other nodes joining to the cluster.
The new cluster will have a token generated for it or you can manually
assign a unique string.  The token for an existing cluster can be determined
by running `rancherd get-token`.

## Joining Nodes

Nodes can be joined to the cluster as the role `server` to add more control
plane nodes or as the role `agent` to add more worker nodes. To join a node
you must have the Rancher server URL (which is by default running on port
`8443`) and the token.

## Node Roles


Rancherd will bootstrap a node with one of the following roles

2. __server__: Joins the cluster as a new control-plane,etcd,worker node
3. __agent__: Joins the cluster as a worker only node.

## Server discovery

It can be quite cumbersome to automate bringing up a clustered system
that requires one bootstrap node.  Also there are more considerations
around load balancing and replacing nodes in a proper production setup.
Rancherd support server discovery based on https://github.com/hashicorp/go-discover.

When using server discovery the `cluster-init` role is not used, only `server`
and `agent`. The `server` URL is also dropped in place of using the `discovery`
key. The `discovery` configuration will be used to dynamically determine what
is the server URL and if the current node should act as the `cluster-init` node.

Example
```yaml
role: server
discovery:
  params:
    # Corresponds to go-discover provider name
    provider: "mdns"
    # All other key/values are parameters corresponding to what 
    # the go-discover provider is expecting
    service: "rancher-server"
  # If this is a new cluster it will wait until 3 server are 
  # available and they all agree on the same cluster-init node
  expectedServers: 3
  # How long servers are remembered for. It is useful for providers
  # that are not consistent in their responses, like mdns.
  serverCacheDuration: 1m
```

More information on how to use the discovery is in the config examples.