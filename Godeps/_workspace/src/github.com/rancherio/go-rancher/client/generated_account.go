package client

const (
	ACCOUNT_TYPE = "account"
)

type Account struct {
	Resource
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    ExternalId string `json:"externalId,omitempty"`
    
    ExternalIdType string `json:"externalIdType,omitempty"`
    
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

type AccountCollection struct {
	Collection
	Data []Account `json:"data,omitempty"`
}

type AccountClient struct {
	rancherClient *RancherClient
}

type AccountOperations interface {
	List(opts *ListOpts) (*AccountCollection, error)
	Create(opts *Account) (*Account, error)
	Update(existing *Account, updates interface{}) (*Account, error)
	ById(id string) (*Account, error)
	Delete(container *Account) error
    ActionActivate (*Account) (*Account, error)
    ActionCreate (*Account) (*Account, error)
    ActionDeactivate (*Account) (*Account, error)
    ActionPurge (*Account) (*Account, error)
    ActionRemove (*Account) (*Account, error)
    ActionRestore (*Account) (*Account, error)
    ActionUpdate (*Account) (*Account, error)
}

func newAccountClient(rancherClient *RancherClient) *AccountClient {
	return &AccountClient{
		rancherClient: rancherClient,
	}
}

func (c *AccountClient) Create(container *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doCreate(ACCOUNT_TYPE, container, resp)
	return resp, err
}

func (c *AccountClient) Update(existing *Account, updates interface{}) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doUpdate(ACCOUNT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *AccountClient) List(opts *ListOpts) (*AccountCollection, error) {
	resp := &AccountCollection{}
	err := c.rancherClient.doList(ACCOUNT_TYPE, opts, resp)
	return resp, err
}

func (c *AccountClient) ById(id string) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doById(ACCOUNT_TYPE, id, resp)
	return resp, err
}

func (c *AccountClient) Delete(container *Account) error {
	return c.rancherClient.doResourceDelete(ACCOUNT_TYPE, &container.Resource)
}

func (c *AccountClient) ActionActivate(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *AccountClient) ActionCreate(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *AccountClient) ActionDeactivate(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *AccountClient) ActionPurge(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *AccountClient) ActionRemove(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *AccountClient) ActionRestore(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *AccountClient) ActionUpdate(resource *Account) (*Account, error) {
	resp := &Account{}
	err := c.rancherClient.doEmptyAction(ACCOUNT_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
