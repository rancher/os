package client

const (
	VIRTUALBOX_CONFIG_TYPE = "virtualboxConfig"
)

type VirtualboxConfig struct {
	Resource
    
    Boot2dockerUrl string `json:"boot2dockerUrl,omitempty"`
    
    DiskSize string `json:"diskSize,omitempty"`
    
    Memory string `json:"memory,omitempty"`
    
}

type VirtualboxConfigCollection struct {
	Collection
	Data []VirtualboxConfig `json:"data,omitempty"`
}

type VirtualboxConfigClient struct {
	rancherClient *RancherClient
}

type VirtualboxConfigOperations interface {
	List(opts *ListOpts) (*VirtualboxConfigCollection, error)
	Create(opts *VirtualboxConfig) (*VirtualboxConfig, error)
	Update(existing *VirtualboxConfig, updates interface{}) (*VirtualboxConfig, error)
	ById(id string) (*VirtualboxConfig, error)
	Delete(container *VirtualboxConfig) error
}

func newVirtualboxConfigClient(rancherClient *RancherClient) *VirtualboxConfigClient {
	return &VirtualboxConfigClient{
		rancherClient: rancherClient,
	}
}

func (c *VirtualboxConfigClient) Create(container *VirtualboxConfig) (*VirtualboxConfig, error) {
	resp := &VirtualboxConfig{}
	err := c.rancherClient.doCreate(VIRTUALBOX_CONFIG_TYPE, container, resp)
	return resp, err
}

func (c *VirtualboxConfigClient) Update(existing *VirtualboxConfig, updates interface{}) (*VirtualboxConfig, error) {
	resp := &VirtualboxConfig{}
	err := c.rancherClient.doUpdate(VIRTUALBOX_CONFIG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *VirtualboxConfigClient) List(opts *ListOpts) (*VirtualboxConfigCollection, error) {
	resp := &VirtualboxConfigCollection{}
	err := c.rancherClient.doList(VIRTUALBOX_CONFIG_TYPE, opts, resp)
	return resp, err
}

func (c *VirtualboxConfigClient) ById(id string) (*VirtualboxConfig, error) {
	resp := &VirtualboxConfig{}
	err := c.rancherClient.doById(VIRTUALBOX_CONFIG_TYPE, id, resp)
	return resp, err
}

func (c *VirtualboxConfigClient) Delete(container *VirtualboxConfig) error {
	return c.rancherClient.doResourceDelete(VIRTUALBOX_CONFIG_TYPE, &container.Resource)
}
