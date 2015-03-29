package client

const (
	GLOBAL_LOAD_BALANCER_TYPE = "globalLoadBalancer"
)

type GlobalLoadBalancer struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    GlobalLoadBalancerHealthCheck []interface{} `json:"globalLoadBalancerHealthCheck,omitempty"`
    
    GlobalLoadBalancerPolicy []interface{} `json:"globalLoadBalancerPolicy,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type GlobalLoadBalancerCollection struct {
	Collection
	Data []GlobalLoadBalancer `json:"data,omitempty"`
}

type GlobalLoadBalancerClient struct {
	rancherClient *RancherClient
}

type GlobalLoadBalancerOperations interface {
	List(opts *ListOpts) (*GlobalLoadBalancerCollection, error)
	Create(opts *GlobalLoadBalancer) (*GlobalLoadBalancer, error)
	Update(existing *GlobalLoadBalancer, updates interface{}) (*GlobalLoadBalancer, error)
	ById(id string) (*GlobalLoadBalancer, error)
	Delete(container *GlobalLoadBalancer) error
    ActionCreate (*GlobalLoadBalancer) (*GlobalLoadBalancer, error)
    ActionRemove (*GlobalLoadBalancer) (*GlobalLoadBalancer, error)
}

func newGlobalLoadBalancerClient(rancherClient *RancherClient) *GlobalLoadBalancerClient {
	return &GlobalLoadBalancerClient{
		rancherClient: rancherClient,
	}
}

func (c *GlobalLoadBalancerClient) Create(container *GlobalLoadBalancer) (*GlobalLoadBalancer, error) {
	resp := &GlobalLoadBalancer{}
	err := c.rancherClient.doCreate(GLOBAL_LOAD_BALANCER_TYPE, container, resp)
	return resp, err
}

func (c *GlobalLoadBalancerClient) Update(existing *GlobalLoadBalancer, updates interface{}) (*GlobalLoadBalancer, error) {
	resp := &GlobalLoadBalancer{}
	err := c.rancherClient.doUpdate(GLOBAL_LOAD_BALANCER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *GlobalLoadBalancerClient) List(opts *ListOpts) (*GlobalLoadBalancerCollection, error) {
	resp := &GlobalLoadBalancerCollection{}
	err := c.rancherClient.doList(GLOBAL_LOAD_BALANCER_TYPE, opts, resp)
	return resp, err
}

func (c *GlobalLoadBalancerClient) ById(id string) (*GlobalLoadBalancer, error) {
	resp := &GlobalLoadBalancer{}
	err := c.rancherClient.doById(GLOBAL_LOAD_BALANCER_TYPE, id, resp)
	return resp, err
}

func (c *GlobalLoadBalancerClient) Delete(container *GlobalLoadBalancer) error {
	return c.rancherClient.doResourceDelete(GLOBAL_LOAD_BALANCER_TYPE, &container.Resource)
}

func (c *GlobalLoadBalancerClient) ActionCreate(resource *GlobalLoadBalancer) (*GlobalLoadBalancer, error) {
	resp := &GlobalLoadBalancer{}
	err := c.rancherClient.doEmptyAction(GLOBAL_LOAD_BALANCER_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *GlobalLoadBalancerClient) ActionRemove(resource *GlobalLoadBalancer) (*GlobalLoadBalancer, error) {
	resp := &GlobalLoadBalancer{}
	err := c.rancherClient.doEmptyAction(GLOBAL_LOAD_BALANCER_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}
