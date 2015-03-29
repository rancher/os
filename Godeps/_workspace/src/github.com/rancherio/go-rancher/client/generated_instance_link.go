package client

const (
	INSTANCE_LINK_TYPE = "instanceLink"
)

type InstanceLink struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    InstanceId string `json:"instanceId,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    LinkName string `json:"linkName,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    Ports []interface{} `json:"ports,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    TargetInstanceId string `json:"targetInstanceId,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type InstanceLinkCollection struct {
	Collection
	Data []InstanceLink `json:"data,omitempty"`
}

type InstanceLinkClient struct {
	rancherClient *RancherClient
}

type InstanceLinkOperations interface {
	List(opts *ListOpts) (*InstanceLinkCollection, error)
	Create(opts *InstanceLink) (*InstanceLink, error)
	Update(existing *InstanceLink, updates interface{}) (*InstanceLink, error)
	ById(id string) (*InstanceLink, error)
	Delete(container *InstanceLink) error
    ActionActivate (*InstanceLink) (*InstanceLink, error)
    ActionCreate (*InstanceLink) (*InstanceLink, error)
    ActionDeactivate (*InstanceLink) (*InstanceLink, error)
    ActionPurge (*InstanceLink) (*InstanceLink, error)
    ActionRemove (*InstanceLink) (*InstanceLink, error)
    ActionRestore (*InstanceLink) (*InstanceLink, error)
    ActionUpdate (*InstanceLink) (*InstanceLink, error)
}

func newInstanceLinkClient(rancherClient *RancherClient) *InstanceLinkClient {
	return &InstanceLinkClient{
		rancherClient: rancherClient,
	}
}

func (c *InstanceLinkClient) Create(container *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doCreate(INSTANCE_LINK_TYPE, container, resp)
	return resp, err
}

func (c *InstanceLinkClient) Update(existing *InstanceLink, updates interface{}) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doUpdate(INSTANCE_LINK_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *InstanceLinkClient) List(opts *ListOpts) (*InstanceLinkCollection, error) {
	resp := &InstanceLinkCollection{}
	err := c.rancherClient.doList(INSTANCE_LINK_TYPE, opts, resp)
	return resp, err
}

func (c *InstanceLinkClient) ById(id string) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doById(INSTANCE_LINK_TYPE, id, resp)
	return resp, err
}

func (c *InstanceLinkClient) Delete(container *InstanceLink) error {
	return c.rancherClient.doResourceDelete(INSTANCE_LINK_TYPE, &container.Resource)
}

func (c *InstanceLinkClient) ActionActivate(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *InstanceLinkClient) ActionCreate(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *InstanceLinkClient) ActionDeactivate(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *InstanceLinkClient) ActionPurge(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *InstanceLinkClient) ActionRemove(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *InstanceLinkClient) ActionRestore(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *InstanceLinkClient) ActionUpdate(resource *InstanceLink) (*InstanceLink, error) {
	resp := &InstanceLink{}
	err := c.rancherClient.doEmptyAction(INSTANCE_LINK_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
