# Configuration Reference

All configuration should come from RancherOS minimal `cloud-init`.
Below is a reference of supported configuration.  It is important
that the config always starts with `#cloud-config`

```yaml
#cloud-config

# Add additional users or set the password/ssh keys for root
users:
- name: "bar"
  passwd: "foo"
  groups: "users"
  ssh_authorized_keys:
  - faaapploo

# Assigns these keys to the first user in users or root if there
# is none
ssh_authorized_keys:
  - asdd

# Run these commands once the system has fully booted
runcmd:
- foo
 
# Hostname to assign
hostname: "bar"

# Write arbitrary files
write_files:
- encoding: b64
  content: CiMgVGhpcyBmaWxlIGNvbnRyb2xzIHRoZSBzdGF0ZSBvZiBTRUxpbnV4
  path: /foo/bar
  permissions: "0644"
  owner: "bar"

# Rancherd configuration
rancherd:
  ########################################################
  # The below parameters apply to server role that first #
  # initializes the cluster                              #
  ########################################################

  # The Kubernetes version to be installed. This must be a k3s or RKE2 version
  # v1.21 or newer. k3s and RKE2 versions always have a `k3s` or `rke2` in the
  # version string.
  # Valid versions are
  # k3s: curl -sL https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/release-v2.6/data/data.json | jq -r '.k3s.releases[].version'
  # RKE2: curl -sL https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/release-v2.6/data/data.json | jq -r '.rke2.releases[].version'
  kubernetesVersion: v1.22.2+k3s1

  # The Rancher version to be installed or a channel "latest" or "stable"
  rancherVersion: v2.6.0

  # Values set on the Rancher Helm chart. Refer to
  # https://github.com/rancher/rancher/blob/release/v2.6/chart/values.yaml
  # for possible values.
  rancherValues:
    # Below are the default values set

    # Multi-Cluster Management is disabled by default, change to multi-cluster-management=true to enable
    features: multi-cluster-management=false
    # The Rancher UI will run on the host port 8443 by default. Set to 0 to disable
    # and instead use ingress.enabled=true to route traffic through ingress
    hostPort: 8443
    # Accessing ingress is disabled by default.
    ingress:
      enabled: false
    # Don't create a default admin password
    noDefaultAdmin: true
    # The negative value means it will up to that many replicas if there are
    # at least that many nodes available.  For example, if you have 2 nodes and
    # `replicas` is `-3` then 2 replicas will run.  Once you add a third node
    # a then 3 replicas will run
    replicas: -3
    # External TLS is assumed
    tls: external


  # Addition SANs (hostnames) to be added to the generated TLS certificate that
  # served on port 6443.
  tlsSans:
    - additionalhostname.example.com

  # Kubernetes resources that will be created once Rancher is bootstrapped
  resources:
    - kind: ConfigMap
      apiVersion: v1
      metadata:
        name: random
      data:
        key: value

  # Contents of the registries.yaml that will be used by k3s/RKE2. The structure
  # is documented at https://rancher.com/docs/k3s/latest/en/installation/private-registry/
  registries: {}

  # The default registry used for all Rancher container images. For more information
  # refer to https://rancher.com/docs/rancher/v2.6/en/admin-settings/config-private-registry/
  systemDefaultRegistry: someprefix.example.com:5000

  # Advanced: The system agent installer image used for Kubernetes
  runtimeInstallerImage: ...

  # Advanced: The system agent installer image used for Rancher
  rancherInstallerImage: ...

  ###########################################
  # The below parameters apply to all roles #
  ###########################################

  # Generic commands to run before bootstrapping the node.
  preInstructions:
    - name: something
      # This image will be extracted to a temporary folder and
      # set as the current working dir. The command will not run
      # contained or chrooted, this is only a way to copy assets
      # to the host. This is parameter is optional
      image: custom/image:1.1.1
      # Environment variables to set
      env:
        - FOO=BAR
      # Program arguments
      args:
        - arg1
        - arg2
      # Command to run
      command: /bin/dosomething
      # Save output to /var/lib/rancher/rancherd/plan/plan-output.json
      saveOutput: false

  # Generic commands to run after bootstrapping the node.
  postInstructions:
    - name: something
      env:
        - FOO=BAR
      args:
        - arg1
        - arg2
      command: /bin/dosomething
      saveOutput: false

  # The URL to Rancher to join a node. If you have disabled the hostPort and configured
  # TLS then this will be the server you have setup.
  server: https://myserver.example.com:8443

  # A shared secret to join nodes to the cluster
  token: sometoken

  # Instead of setting the server parameter above the server value can be dynamically
  # determined from cloud provider metadata. This is powered by https://github.com/hashicorp/go-discover.
  # Discovery requires that the hostPort is not disabled.
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

  # The role of this node.  Every cluster must start with one node as role=cluster-init.
  # After that nodes can be joined using the server role for control-plane nodes and
  # agent role for worker only nodes.  The server/agent terms correspond to the server/agent
  # terms in k3s and RKE2
  role: cluster-init,server,agent
  # The Kubernetes node name that will be set
  nodeName: custom-hostname
  # The IP address that will be set in Kubernetes for this node
  address: 123.123.123.123
  # The internal IP address that will be used for this node
  internalAddress: 123.123.123.124
  # Taints to apply to this node upon creation
  taints:
    - dedicated=special-user:NoSchedule
  # Labels to apply to this node upon creation
  labels:
    - key=value

  # Advanced: Arbitrary configuration that will be placed in /etc/rancher/k3s/config.yaml.d/40-rancherd.yaml
  # or /etc/rancher/rke2/config.yaml.d/40-rancherd.yaml
  extraConfig: {}
```