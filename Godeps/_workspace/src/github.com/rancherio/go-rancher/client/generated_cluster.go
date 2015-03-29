package client

const (
	CLUSTER_TYPE = "cluster"
)

type Cluster struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    AgentId string `json:"agentId,omitempty"`
    
    ApiProxy string `json:"apiProxy,omitempty"`
    
    CertificateId string `json:"certificateId,omitempty"`
    
    ComputeTotal int `json:"computeTotal,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    DiscoverySpec string `json:"discoverySpec,omitempty"`
    
    Info interface{} `json:"info,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    PhysicalHostId string `json:"physicalHostId,omitempty"`
    
    Port int `json:"port,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type ClusterCollection struct {
	Collection
	Data []Cluster `json:"data,omitempty"`
}

type ClusterClient struct {
	rancherClient *RancherClient
}

type ClusterOperations interface {
	List(opts *ListOpts) (*ClusterCollection, error)
	Create(opts *Cluster) (*Cluster, error)
	Update(existing *Cluster, updates interface{}) (*Cluster, error)
	ById(id string) (*Cluster, error)
	Delete(container *Cluster) error
}

func newClusterClient(rancherClient *RancherClient) *ClusterClient {
	return &ClusterClient{
		rancherClient: rancherClient,
	}
}

func (c *ClusterClient) Create(container *Cluster) (*Cluster, error) {
	resp := &Cluster{}
	err := c.rancherClient.doCreate(CLUSTER_TYPE, container, resp)
	return resp, err
}

func (c *ClusterClient) Update(existing *Cluster, updates interface{}) (*Cluster, error) {
	resp := &Cluster{}
	err := c.rancherClient.doUpdate(CLUSTER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ClusterClient) List(opts *ListOpts) (*ClusterCollection, error) {
	resp := &ClusterCollection{}
	err := c.rancherClient.doList(CLUSTER_TYPE, opts, resp)
	return resp, err
}

func (c *ClusterClient) ById(id string) (*Cluster, error) {
	resp := &Cluster{}
	err := c.rancherClient.doById(CLUSTER_TYPE, id, resp)
	return resp, err
}

func (c *ClusterClient) Delete(container *Cluster) error {
	return c.rancherClient.doResourceDelete(CLUSTER_TYPE, &container.Resource)
}
