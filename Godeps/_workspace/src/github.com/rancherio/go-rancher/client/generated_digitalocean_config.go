package client

const (
	DIGITALOCEAN_CONFIG_TYPE = "digitaloceanConfig"
)

type DigitaloceanConfig struct {
	Resource
    
    AccessToken string `json:"accessToken,omitempty"`
    
    Image string `json:"image,omitempty"`
    
    Region string `json:"region,omitempty"`
    
    Size string `json:"size,omitempty"`
    
}

type DigitaloceanConfigCollection struct {
	Collection
	Data []DigitaloceanConfig `json:"data,omitempty"`
}

type DigitaloceanConfigClient struct {
	rancherClient *RancherClient
}

type DigitaloceanConfigOperations interface {
	List(opts *ListOpts) (*DigitaloceanConfigCollection, error)
	Create(opts *DigitaloceanConfig) (*DigitaloceanConfig, error)
	Update(existing *DigitaloceanConfig, updates interface{}) (*DigitaloceanConfig, error)
	ById(id string) (*DigitaloceanConfig, error)
	Delete(container *DigitaloceanConfig) error
}

func newDigitaloceanConfigClient(rancherClient *RancherClient) *DigitaloceanConfigClient {
	return &DigitaloceanConfigClient{
		rancherClient: rancherClient,
	}
}

func (c *DigitaloceanConfigClient) Create(container *DigitaloceanConfig) (*DigitaloceanConfig, error) {
	resp := &DigitaloceanConfig{}
	err := c.rancherClient.doCreate(DIGITALOCEAN_CONFIG_TYPE, container, resp)
	return resp, err
}

func (c *DigitaloceanConfigClient) Update(existing *DigitaloceanConfig, updates interface{}) (*DigitaloceanConfig, error) {
	resp := &DigitaloceanConfig{}
	err := c.rancherClient.doUpdate(DIGITALOCEAN_CONFIG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *DigitaloceanConfigClient) List(opts *ListOpts) (*DigitaloceanConfigCollection, error) {
	resp := &DigitaloceanConfigCollection{}
	err := c.rancherClient.doList(DIGITALOCEAN_CONFIG_TYPE, opts, resp)
	return resp, err
}

func (c *DigitaloceanConfigClient) ById(id string) (*DigitaloceanConfig, error) {
	resp := &DigitaloceanConfig{}
	err := c.rancherClient.doById(DIGITALOCEAN_CONFIG_TYPE, id, resp)
	return resp, err
}

func (c *DigitaloceanConfigClient) Delete(container *DigitaloceanConfig) error {
	return c.rancherClient.doResourceDelete(DIGITALOCEAN_CONFIG_TYPE, &container.Resource)
}
