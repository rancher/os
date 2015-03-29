package client

const (
	API_KEY_TYPE = "apiKey"
)

type ApiKey struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    PublicValue string `json:"publicValue,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    SecretValue string `json:"secretValue,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type ApiKeyCollection struct {
	Collection
	Data []ApiKey `json:"data,omitempty"`
}

type ApiKeyClient struct {
	rancherClient *RancherClient
}

type ApiKeyOperations interface {
	List(opts *ListOpts) (*ApiKeyCollection, error)
	Create(opts *ApiKey) (*ApiKey, error)
	Update(existing *ApiKey, updates interface{}) (*ApiKey, error)
	ById(id string) (*ApiKey, error)
	Delete(container *ApiKey) error
}

func newApiKeyClient(rancherClient *RancherClient) *ApiKeyClient {
	return &ApiKeyClient{
		rancherClient: rancherClient,
	}
}

func (c *ApiKeyClient) Create(container *ApiKey) (*ApiKey, error) {
	resp := &ApiKey{}
	err := c.rancherClient.doCreate(API_KEY_TYPE, container, resp)
	return resp, err
}

func (c *ApiKeyClient) Update(existing *ApiKey, updates interface{}) (*ApiKey, error) {
	resp := &ApiKey{}
	err := c.rancherClient.doUpdate(API_KEY_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ApiKeyClient) List(opts *ListOpts) (*ApiKeyCollection, error) {
	resp := &ApiKeyCollection{}
	err := c.rancherClient.doList(API_KEY_TYPE, opts, resp)
	return resp, err
}

func (c *ApiKeyClient) ById(id string) (*ApiKey, error) {
	resp := &ApiKey{}
	err := c.rancherClient.doById(API_KEY_TYPE, id, resp)
	return resp, err
}

func (c *ApiKeyClient) Delete(container *ApiKey) error {
	return c.rancherClient.doResourceDelete(API_KEY_TYPE, &container.Resource)
}
