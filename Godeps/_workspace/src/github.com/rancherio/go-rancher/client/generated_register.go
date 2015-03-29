package client

const (
	REGISTER_TYPE = "register"
)

type Register struct {
	Resource
    
    AccessKey string `json:"accessKey,omitempty"`
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Key string `json:"key,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    SecretKey string `json:"secretKey,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type RegisterCollection struct {
	Collection
	Data []Register `json:"data,omitempty"`
}

type RegisterClient struct {
	rancherClient *RancherClient
}

type RegisterOperations interface {
	List(opts *ListOpts) (*RegisterCollection, error)
	Create(opts *Register) (*Register, error)
	Update(existing *Register, updates interface{}) (*Register, error)
	ById(id string) (*Register, error)
	Delete(container *Register) error
}

func newRegisterClient(rancherClient *RancherClient) *RegisterClient {
	return &RegisterClient{
		rancherClient: rancherClient,
	}
}

func (c *RegisterClient) Create(container *Register) (*Register, error) {
	resp := &Register{}
	err := c.rancherClient.doCreate(REGISTER_TYPE, container, resp)
	return resp, err
}

func (c *RegisterClient) Update(existing *Register, updates interface{}) (*Register, error) {
	resp := &Register{}
	err := c.rancherClient.doUpdate(REGISTER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *RegisterClient) List(opts *ListOpts) (*RegisterCollection, error) {
	resp := &RegisterCollection{}
	err := c.rancherClient.doList(REGISTER_TYPE, opts, resp)
	return resp, err
}

func (c *RegisterClient) ById(id string) (*Register, error) {
	resp := &Register{}
	err := c.rancherClient.doById(REGISTER_TYPE, id, resp)
	return resp, err
}

func (c *RegisterClient) Delete(container *Register) error {
	return c.rancherClient.doResourceDelete(REGISTER_TYPE, &container.Resource)
}
