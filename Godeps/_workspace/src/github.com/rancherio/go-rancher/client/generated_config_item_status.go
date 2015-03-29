package client

const (
	CONFIG_ITEM_STATUS_TYPE = "configItemStatus"
)

type ConfigItemStatus struct {
	Resource
    
    AgentId string `json:"agentId,omitempty"`
    
    AppliedUpdated string `json:"appliedUpdated,omitempty"`
    
    AppliedVersion int `json:"appliedVersion,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RequestedUpdated string `json:"requestedUpdated,omitempty"`
    
    RequestedVersion int `json:"requestedVersion,omitempty"`
    
    SourceVersion string `json:"sourceVersion,omitempty"`
    
}

type ConfigItemStatusCollection struct {
	Collection
	Data []ConfigItemStatus `json:"data,omitempty"`
}

type ConfigItemStatusClient struct {
	rancherClient *RancherClient
}

type ConfigItemStatusOperations interface {
	List(opts *ListOpts) (*ConfigItemStatusCollection, error)
	Create(opts *ConfigItemStatus) (*ConfigItemStatus, error)
	Update(existing *ConfigItemStatus, updates interface{}) (*ConfigItemStatus, error)
	ById(id string) (*ConfigItemStatus, error)
	Delete(container *ConfigItemStatus) error
}

func newConfigItemStatusClient(rancherClient *RancherClient) *ConfigItemStatusClient {
	return &ConfigItemStatusClient{
		rancherClient: rancherClient,
	}
}

func (c *ConfigItemStatusClient) Create(container *ConfigItemStatus) (*ConfigItemStatus, error) {
	resp := &ConfigItemStatus{}
	err := c.rancherClient.doCreate(CONFIG_ITEM_STATUS_TYPE, container, resp)
	return resp, err
}

func (c *ConfigItemStatusClient) Update(existing *ConfigItemStatus, updates interface{}) (*ConfigItemStatus, error) {
	resp := &ConfigItemStatus{}
	err := c.rancherClient.doUpdate(CONFIG_ITEM_STATUS_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ConfigItemStatusClient) List(opts *ListOpts) (*ConfigItemStatusCollection, error) {
	resp := &ConfigItemStatusCollection{}
	err := c.rancherClient.doList(CONFIG_ITEM_STATUS_TYPE, opts, resp)
	return resp, err
}

func (c *ConfigItemStatusClient) ById(id string) (*ConfigItemStatus, error) {
	resp := &ConfigItemStatus{}
	err := c.rancherClient.doById(CONFIG_ITEM_STATUS_TYPE, id, resp)
	return resp, err
}

func (c *ConfigItemStatusClient) Delete(container *ConfigItemStatus) error {
	return c.rancherClient.doResourceDelete(CONFIG_ITEM_STATUS_TYPE, &container.Resource)
}
