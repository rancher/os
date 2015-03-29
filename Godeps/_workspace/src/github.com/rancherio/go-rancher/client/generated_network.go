package client

const (
	NETWORK_TYPE = "network"
)

type Network struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
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

type NetworkCollection struct {
	Collection
	Data []Network `json:"data,omitempty"`
}

type NetworkClient struct {
	rancherClient *RancherClient
}

type NetworkOperations interface {
	List(opts *ListOpts) (*NetworkCollection, error)
	Create(opts *Network) (*Network, error)
	Update(existing *Network, updates interface{}) (*Network, error)
	ById(id string) (*Network, error)
	Delete(container *Network) error
    ActionActivate (*Network) (*Network, error)
    ActionCreate (*Network) (*Network, error)
    ActionDeactivate (*Network) (*Network, error)
    ActionPurge (*Network) (*Network, error)
    ActionRemove (*Network) (*Network, error)
    ActionRestore (*Network) (*Network, error)
    ActionUpdate (*Network) (*Network, error)
}

func newNetworkClient(rancherClient *RancherClient) *NetworkClient {
	return &NetworkClient{
		rancherClient: rancherClient,
	}
}

func (c *NetworkClient) Create(container *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doCreate(NETWORK_TYPE, container, resp)
	return resp, err
}

func (c *NetworkClient) Update(existing *Network, updates interface{}) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doUpdate(NETWORK_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *NetworkClient) List(opts *ListOpts) (*NetworkCollection, error) {
	resp := &NetworkCollection{}
	err := c.rancherClient.doList(NETWORK_TYPE, opts, resp)
	return resp, err
}

func (c *NetworkClient) ById(id string) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doById(NETWORK_TYPE, id, resp)
	return resp, err
}

func (c *NetworkClient) Delete(container *Network) error {
	return c.rancherClient.doResourceDelete(NETWORK_TYPE, &container.Resource)
}

func (c *NetworkClient) ActionActivate(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *NetworkClient) ActionCreate(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *NetworkClient) ActionDeactivate(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *NetworkClient) ActionPurge(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *NetworkClient) ActionRemove(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *NetworkClient) ActionRestore(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *NetworkClient) ActionUpdate(resource *Network) (*Network, error) {
	resp := &Network{}
	err := c.rancherClient.doEmptyAction(NETWORK_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
