package client

const (
	LOAD_BALANCER_CONFIG_TYPE = "loadBalancerConfig"
)

type LoadBalancerConfig struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    AppCookieStickinessPolicy LoadBalancerAppCookieStickinessPolicy `json:"appCookieStickinessPolicy,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    HealthCheck LoadBalancerHealthCheck `json:"healthCheck,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    LbCookieStickinessPolicy LoadBalancerCookieStickinessPolicy `json:"lbCookieStickinessPolicy,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type LoadBalancerConfigCollection struct {
	Collection
	Data []LoadBalancerConfig `json:"data,omitempty"`
}

type LoadBalancerConfigClient struct {
	rancherClient *RancherClient
}

type LoadBalancerConfigOperations interface {
	List(opts *ListOpts) (*LoadBalancerConfigCollection, error)
	Create(opts *LoadBalancerConfig) (*LoadBalancerConfig, error)
	Update(existing *LoadBalancerConfig, updates interface{}) (*LoadBalancerConfig, error)
	ById(id string) (*LoadBalancerConfig, error)
	Delete(container *LoadBalancerConfig) error
    ActionCreate (*LoadBalancerConfig) (*LoadBalancerConfig, error)
    ActionRemove (*LoadBalancerConfig) (*LoadBalancerConfig, error)
    ActionUpdate (*LoadBalancerConfig) (*LoadBalancerConfig, error)
}

func newLoadBalancerConfigClient(rancherClient *RancherClient) *LoadBalancerConfigClient {
	return &LoadBalancerConfigClient{
		rancherClient: rancherClient,
	}
}

func (c *LoadBalancerConfigClient) Create(container *LoadBalancerConfig) (*LoadBalancerConfig, error) {
	resp := &LoadBalancerConfig{}
	err := c.rancherClient.doCreate(LOAD_BALANCER_CONFIG_TYPE, container, resp)
	return resp, err
}

func (c *LoadBalancerConfigClient) Update(existing *LoadBalancerConfig, updates interface{}) (*LoadBalancerConfig, error) {
	resp := &LoadBalancerConfig{}
	err := c.rancherClient.doUpdate(LOAD_BALANCER_CONFIG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LoadBalancerConfigClient) List(opts *ListOpts) (*LoadBalancerConfigCollection, error) {
	resp := &LoadBalancerConfigCollection{}
	err := c.rancherClient.doList(LOAD_BALANCER_CONFIG_TYPE, opts, resp)
	return resp, err
}

func (c *LoadBalancerConfigClient) ById(id string) (*LoadBalancerConfig, error) {
	resp := &LoadBalancerConfig{}
	err := c.rancherClient.doById(LOAD_BALANCER_CONFIG_TYPE, id, resp)
	return resp, err
}

func (c *LoadBalancerConfigClient) Delete(container *LoadBalancerConfig) error {
	return c.rancherClient.doResourceDelete(LOAD_BALANCER_CONFIG_TYPE, &container.Resource)
}

func (c *LoadBalancerConfigClient) ActionCreate(resource *LoadBalancerConfig) (*LoadBalancerConfig, error) {
	resp := &LoadBalancerConfig{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_CONFIG_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *LoadBalancerConfigClient) ActionRemove(resource *LoadBalancerConfig) (*LoadBalancerConfig, error) {
	resp := &LoadBalancerConfig{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_CONFIG_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *LoadBalancerConfigClient) ActionUpdate(resource *LoadBalancerConfig) (*LoadBalancerConfig, error) {
	resp := &LoadBalancerConfig{}
	err := c.rancherClient.doEmptyAction(LOAD_BALANCER_CONFIG_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
