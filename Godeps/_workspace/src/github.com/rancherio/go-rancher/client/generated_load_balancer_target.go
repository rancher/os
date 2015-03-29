package client

const (
	LOAD_BALANCER_TARGET_TYPE = "loadBalancerTarget"
)

type LoadBalancerTarget struct {
	Resource
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    InstanceId string `json:"instanceId,omitempty"`
    
    IpAddress string `json:"ipAddress,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    LoadBalancerId string `json:"loadBalancerId,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type LoadBalancerTargetCollection struct {
	Collection
	Data []LoadBalancerTarget `json:"data,omitempty"`
}

type LoadBalancerTargetClient struct {
	rancherClient *RancherClient
}

type LoadBalancerTargetOperations interface {
	List(opts *ListOpts) (*LoadBalancerTargetCollection, error)
	Create(opts *LoadBalancerTarget) (*LoadBalancerTarget, error)
	Update(existing *LoadBalancerTarget, updates interface{}) (*LoadBalancerTarget, error)
	ById(id string) (*LoadBalancerTarget, error)
	Delete(container *LoadBalancerTarget) error
    ActionCreate (*LoadBalancerTarget) (*LoadBalancerTarget, error)
    ActionRemove (*LoadBalancerTarget) (*LoadBalancerTarget, error)
}

func newLoadBalancerTargetClient(rancherClient *RancherClient) *LoadBalancerTargetClient {
	return &LoadBalancerTargetClient{
		rancherClient: rancherClient,
	}
}

func (c *LoadBalancerTargetClient) Create(container *LoadBalancerTarget) (*LoadBalancerTarget, error) {
	resp := &LoadBalancerTarget{}
	err := c.rancherClient.doCreate(LOAD_BALANCER_TARGET_TYPE, container, resp)
	return resp, err
}

func (c *LoadBalancerTargetClient) Update(existing *LoadBalancerTarget, updates interface{}) (*LoadBalancerTarget, error) {
	resp := &LoadBalancerTarget{}
	err := c.rancherClient.doUpdate(LOAD_BALANCER_TARGET_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LoadBalancerTargetClient) List(opts *ListOpts) (*LoadBalancerTargetCollection, error) {
	resp := &LoadBalancerTargetCollection{}
	err := c.rancherClient.doList(LOAD_BALANCER_TARGET_TYPE, opts, resp)
	return resp, err
}

func (c *LoadBalancerTargetClient) ById(id string) (*LoadBalancerTarget, error) {
	resp := &LoadBalancerTarget{}
	err := c.rancherClient.doById(LOAD_BALANCER_TARGET_TYPE, id, resp)
	return resp, err
}

func (c *LoadBalancerTargetClient) Delete(container *LoadBalancerTarget) error {
	return c.rancherClient.doResourceDelete(LOAD_BALANCER_TARGET_TYPE, &container.Resource)
}

func (c *LoadBalancerTargetClient) ActionCreate(resource *LoadBalancerTarget) (*LoadBalancerTarget, error) {
	resp := &LoadBalancerTarget{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_TARGET_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *LoadBalancerTargetClient) ActionRemove(resource *LoadBalancerTarget) (*LoadBalancerTarget, error) {
	resp := &LoadBalancerTarget{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_TARGET_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}
