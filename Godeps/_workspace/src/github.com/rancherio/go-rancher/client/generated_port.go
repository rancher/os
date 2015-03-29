package client

const (
	PORT_TYPE = "port"
)

type Port struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    InstanceId string `json:"instanceId,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    PrivateIpAddressId string `json:"privateIpAddressId,omitempty"`
    
    PrivatePort int `json:"privatePort,omitempty"`
    
    Protocol string `json:"protocol,omitempty"`
    
    PublicIpAddressId string `json:"publicIpAddressId,omitempty"`
    
    PublicPort int `json:"publicPort,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type PortCollection struct {
	Collection
	Data []Port `json:"data,omitempty"`
}

type PortClient struct {
	rancherClient *RancherClient
}

type PortOperations interface {
	List(opts *ListOpts) (*PortCollection, error)
	Create(opts *Port) (*Port, error)
	Update(existing *Port, updates interface{}) (*Port, error)
	ById(id string) (*Port, error)
	Delete(container *Port) error
    ActionActivate (*Port) (*Port, error)
    ActionCreate (*Port) (*Port, error)
    ActionDeactivate (*Port) (*Port, error)
    ActionPurge (*Port) (*Port, error)
    ActionRemove (*Port) (*Port, error)
    ActionRestore (*Port) (*Port, error)
    ActionUpdate (*Port) (*Port, error)
}

func newPortClient(rancherClient *RancherClient) *PortClient {
	return &PortClient{
		rancherClient: rancherClient,
	}
}

func (c *PortClient) Create(container *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doCreate(PORT_TYPE, container, resp)
	return resp, err
}

func (c *PortClient) Update(existing *Port, updates interface{}) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doUpdate(PORT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PortClient) List(opts *ListOpts) (*PortCollection, error) {
	resp := &PortCollection{}
	err := c.rancherClient.doList(PORT_TYPE, opts, resp)
	return resp, err
}

func (c *PortClient) ById(id string) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doById(PORT_TYPE, id, resp)
	return resp, err
}

func (c *PortClient) Delete(container *Port) error {
	return c.rancherClient.doResourceDelete(PORT_TYPE, &container.Resource)
}

func (c *PortClient) ActionActivate(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *PortClient) ActionCreate(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *PortClient) ActionDeactivate(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *PortClient) ActionPurge(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *PortClient) ActionRemove(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *PortClient) ActionRestore(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *PortClient) ActionUpdate(resource *Port) (*Port, error) {
	resp := &Port{}
	err := c.rancherClient.doEmptyAction(PORT_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
