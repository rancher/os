package client

const (
	LOAD_BALANCER_POLICY_TYPE = "loadBalancerPolicy"
)

type LoadBalancerPolicy struct {
	Resource
    
    Name string `json:"name,omitempty"`
    
}

type LoadBalancerPolicyCollection struct {
	Collection
	Data []LoadBalancerPolicy `json:"data,omitempty"`
}

type LoadBalancerPolicyClient struct {
	rancherClient *RancherClient
}

type LoadBalancerPolicyOperations interface {
	List(opts *ListOpts) (*LoadBalancerPolicyCollection, error)
	Create(opts *LoadBalancerPolicy) (*LoadBalancerPolicy, error)
	Update(existing *LoadBalancerPolicy, updates interface{}) (*LoadBalancerPolicy, error)
	ById(id string) (*LoadBalancerPolicy, error)
	Delete(container *LoadBalancerPolicy) error
}

func newLoadBalancerPolicyClient(rancherClient *RancherClient) *LoadBalancerPolicyClient {
	return &LoadBalancerPolicyClient{
		rancherClient: rancherClient,
	}
}

func (c *LoadBalancerPolicyClient) Create(container *LoadBalancerPolicy) (*LoadBalancerPolicy, error) {
	resp := &LoadBalancerPolicy{}
	err := c.rancherClient.doCreate(LOAD_BALANCER_POLICY_TYPE, container, resp)
	return resp, err
}

func (c *LoadBalancerPolicyClient) Update(existing *LoadBalancerPolicy, updates interface{}) (*LoadBalancerPolicy, error) {
	resp := &LoadBalancerPolicy{}
	err := c.rancherClient.doUpdate(LOAD_BALANCER_POLICY_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LoadBalancerPolicyClient) List(opts *ListOpts) (*LoadBalancerPolicyCollection, error) {
	resp := &LoadBalancerPolicyCollection{}
	err := c.rancherClient.doList(LOAD_BALANCER_POLICY_TYPE, opts, resp)
	return resp, err
}

func (c *LoadBalancerPolicyClient) ById(id string) (*LoadBalancerPolicy, error) {
	resp := &LoadBalancerPolicy{}
	err := c.rancherClient.doById(LOAD_BALANCER_POLICY_TYPE, id, resp)
	return resp, err
}

func (c *LoadBalancerPolicyClient) Delete(container *LoadBalancerPolicy) error {
	return c.rancherClient.doResourceDelete(LOAD_BALANCER_POLICY_TYPE, &container.Resource)
}
