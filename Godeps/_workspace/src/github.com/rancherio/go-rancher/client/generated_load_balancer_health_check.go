package client

const (
	LOAD_BALANCER_HEALTH_CHECK_TYPE = "loadBalancerHealthCheck"
)

type LoadBalancerHealthCheck struct {
	Resource
    
    HealthyThreshold int `json:"healthyThreshold,omitempty"`
    
    Interval int `json:"interval,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    ResponseTimeout int `json:"responseTimeout,omitempty"`
    
    UnhealthyThreshold int `json:"unhealthyThreshold,omitempty"`
    
    Uri string `json:"uri,omitempty"`
    
}

type LoadBalancerHealthCheckCollection struct {
	Collection
	Data []LoadBalancerHealthCheck `json:"data,omitempty"`
}

type LoadBalancerHealthCheckClient struct {
	rancherClient *RancherClient
}

type LoadBalancerHealthCheckOperations interface {
	List(opts *ListOpts) (*LoadBalancerHealthCheckCollection, error)
	Create(opts *LoadBalancerHealthCheck) (*LoadBalancerHealthCheck, error)
	Update(existing *LoadBalancerHealthCheck, updates interface{}) (*LoadBalancerHealthCheck, error)
	ById(id string) (*LoadBalancerHealthCheck, error)
	Delete(container *LoadBalancerHealthCheck) error
}

func newLoadBalancerHealthCheckClient(rancherClient *RancherClient) *LoadBalancerHealthCheckClient {
	return &LoadBalancerHealthCheckClient{
		rancherClient: rancherClient,
	}
}

func (c *LoadBalancerHealthCheckClient) Create(container *LoadBalancerHealthCheck) (*LoadBalancerHealthCheck, error) {
	resp := &LoadBalancerHealthCheck{}
	err := c.rancherClient.doCreate(LOAD_BALANCER_HEALTH_CHECK_TYPE, container, resp)
	return resp, err
}

func (c *LoadBalancerHealthCheckClient) Update(existing *LoadBalancerHealthCheck, updates interface{}) (*LoadBalancerHealthCheck, error) {
	resp := &LoadBalancerHealthCheck{}
	err := c.rancherClient.doUpdate(LOAD_BALANCER_HEALTH_CHECK_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LoadBalancerHealthCheckClient) List(opts *ListOpts) (*LoadBalancerHealthCheckCollection, error) {
	resp := &LoadBalancerHealthCheckCollection{}
	err := c.rancherClient.doList(LOAD_BALANCER_HEALTH_CHECK_TYPE, opts, resp)
	return resp, err
}

func (c *LoadBalancerHealthCheckClient) ById(id string) (*LoadBalancerHealthCheck, error) {
	resp := &LoadBalancerHealthCheck{}
	err := c.rancherClient.doById(LOAD_BALANCER_HEALTH_CHECK_TYPE, id, resp)
	return resp, err
}

func (c *LoadBalancerHealthCheckClient) Delete(container *LoadBalancerHealthCheck) error {
	return c.rancherClient.doResourceDelete(LOAD_BALANCER_HEALTH_CHECK_TYPE, &container.Resource)
}
