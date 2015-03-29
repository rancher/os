package client

const (
	SET_LOAD_BALANCER_TARGETS_INPUT_TYPE = "setLoadBalancerTargetsInput"
)

type SetLoadBalancerTargetsInput struct {
	Resource
    
    InstanceIds []string `json:"instanceIds,omitempty"`
    
    IpAddresses []string `json:"ipAddresses,omitempty"`
    
}

type SetLoadBalancerTargetsInputCollection struct {
	Collection
	Data []SetLoadBalancerTargetsInput `json:"data,omitempty"`
}

type SetLoadBalancerTargetsInputClient struct {
	rancherClient *RancherClient
}

type SetLoadBalancerTargetsInputOperations interface {
	List(opts *ListOpts) (*SetLoadBalancerTargetsInputCollection, error)
	Create(opts *SetLoadBalancerTargetsInput) (*SetLoadBalancerTargetsInput, error)
	Update(existing *SetLoadBalancerTargetsInput, updates interface{}) (*SetLoadBalancerTargetsInput, error)
	ById(id string) (*SetLoadBalancerTargetsInput, error)
	Delete(container *SetLoadBalancerTargetsInput) error
}

func newSetLoadBalancerTargetsInputClient(rancherClient *RancherClient) *SetLoadBalancerTargetsInputClient {
	return &SetLoadBalancerTargetsInputClient{
		rancherClient: rancherClient,
	}
}

func (c *SetLoadBalancerTargetsInputClient) Create(container *SetLoadBalancerTargetsInput) (*SetLoadBalancerTargetsInput, error) {
	resp := &SetLoadBalancerTargetsInput{}
	err := c.rancherClient.doCreate(SET_LOAD_BALANCER_TARGETS_INPUT_TYPE, container, resp)
	return resp, err
}

func (c *SetLoadBalancerTargetsInputClient) Update(existing *SetLoadBalancerTargetsInput, updates interface{}) (*SetLoadBalancerTargetsInput, error) {
	resp := &SetLoadBalancerTargetsInput{}
	err := c.rancherClient.doUpdate(SET_LOAD_BALANCER_TARGETS_INPUT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *SetLoadBalancerTargetsInputClient) List(opts *ListOpts) (*SetLoadBalancerTargetsInputCollection, error) {
	resp := &SetLoadBalancerTargetsInputCollection{}
	err := c.rancherClient.doList(SET_LOAD_BALANCER_TARGETS_INPUT_TYPE, opts, resp)
	return resp, err
}

func (c *SetLoadBalancerTargetsInputClient) ById(id string) (*SetLoadBalancerTargetsInput, error) {
	resp := &SetLoadBalancerTargetsInput{}
	err := c.rancherClient.doById(SET_LOAD_BALANCER_TARGETS_INPUT_TYPE, id, resp)
	return resp, err
}

func (c *SetLoadBalancerTargetsInputClient) Delete(container *SetLoadBalancerTargetsInput) error {
	return c.rancherClient.doResourceDelete(SET_LOAD_BALANCER_TARGETS_INPUT_TYPE, &container.Resource)
}
