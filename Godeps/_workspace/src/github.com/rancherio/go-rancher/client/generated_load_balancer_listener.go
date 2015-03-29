package client

const (
	LOAD_BALANCER_LISTENER_TYPE = "loadBalancerListener"
)

type LoadBalancerListener struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Algorithm string `json:"algorithm,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    SourcePort int `json:"sourcePort,omitempty"`
    
    SourceProtocol string `json:"sourceProtocol,omitempty"`
    
    State string `json:"state,omitempty"`
    
    TargetPort int `json:"targetPort,omitempty"`
    
    TargetProtocol string `json:"targetProtocol,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type LoadBalancerListenerCollection struct {
	Collection
	Data []LoadBalancerListener `json:"data,omitempty"`
}

type LoadBalancerListenerClient struct {
	rancherClient *RancherClient
}

type LoadBalancerListenerOperations interface {
	List(opts *ListOpts) (*LoadBalancerListenerCollection, error)
	Create(opts *LoadBalancerListener) (*LoadBalancerListener, error)
	Update(existing *LoadBalancerListener, updates interface{}) (*LoadBalancerListener, error)
	ById(id string) (*LoadBalancerListener, error)
	Delete(container *LoadBalancerListener) error
    ActionCreate (*LoadBalancerListener) (*LoadBalancerListener, error)
    ActionRemove (*LoadBalancerListener) (*LoadBalancerListener, error)
}

func newLoadBalancerListenerClient(rancherClient *RancherClient) *LoadBalancerListenerClient {
	return &LoadBalancerListenerClient{
		rancherClient: rancherClient,
	}
}

func (c *LoadBalancerListenerClient) Create(container *LoadBalancerListener) (*LoadBalancerListener, error) {
	resp := &LoadBalancerListener{}
	err := c.rancherClient.doCreate(LOAD_BALANCER_LISTENER_TYPE, container, resp)
	return resp, err
}

func (c *LoadBalancerListenerClient) Update(existing *LoadBalancerListener, updates interface{}) (*LoadBalancerListener, error) {
	resp := &LoadBalancerListener{}
	err := c.rancherClient.doUpdate(LOAD_BALANCER_LISTENER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LoadBalancerListenerClient) List(opts *ListOpts) (*LoadBalancerListenerCollection, error) {
	resp := &LoadBalancerListenerCollection{}
	err := c.rancherClient.doList(LOAD_BALANCER_LISTENER_TYPE, opts, resp)
	return resp, err
}

func (c *LoadBalancerListenerClient) ById(id string) (*LoadBalancerListener, error) {
	resp := &LoadBalancerListener{}
	err := c.rancherClient.doById(LOAD_BALANCER_LISTENER_TYPE, id, resp)
	return resp, err
}

func (c *LoadBalancerListenerClient) Delete(container *LoadBalancerListener) error {
	return c.rancherClient.doResourceDelete(LOAD_BALANCER_LISTENER_TYPE, &container.Resource)
}

func (c *LoadBalancerListenerClient) ActionCreate(resource *LoadBalancerListener) (*LoadBalancerListener, error) {
	resp := &LoadBalancerListener{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_LISTENER_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *LoadBalancerListenerClient) ActionRemove(resource *LoadBalancerListener) (*LoadBalancerListener, error) {
	resp := &LoadBalancerListener{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_LISTENER_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}
