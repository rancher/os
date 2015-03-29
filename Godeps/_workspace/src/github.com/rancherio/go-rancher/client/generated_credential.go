package client

const (
	CREDENTIAL_TYPE = "credential"
)

type Credential struct {
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

type CredentialCollection struct {
	Collection
	Data []Credential `json:"data,omitempty"`
}

type CredentialClient struct {
	rancherClient *RancherClient
}

type CredentialOperations interface {
	List(opts *ListOpts) (*CredentialCollection, error)
	Create(opts *Credential) (*Credential, error)
	Update(existing *Credential, updates interface{}) (*Credential, error)
	ById(id string) (*Credential, error)
	Delete(container *Credential) error
    ActionActivate (*Credential) (*Credential, error)
    ActionCreate (*Credential) (*Credential, error)
    ActionDeactivate (*Credential) (*Credential, error)
    ActionPurge (*Credential) (*Credential, error)
    ActionRemove (*Credential) (*Credential, error)
    ActionRestore (*Credential) (*Credential, error)
    ActionUpdate (*Credential) (*Credential, error)
}

func newCredentialClient(rancherClient *RancherClient) *CredentialClient {
	return &CredentialClient{
		rancherClient: rancherClient,
	}
}

func (c *CredentialClient) Create(container *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doCreate(CREDENTIAL_TYPE, container, resp)
	return resp, err
}

func (c *CredentialClient) Update(existing *Credential, updates interface{}) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doUpdate(CREDENTIAL_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *CredentialClient) List(opts *ListOpts) (*CredentialCollection, error) {
	resp := &CredentialCollection{}
	err := c.rancherClient.doList(CREDENTIAL_TYPE, opts, resp)
	return resp, err
}

func (c *CredentialClient) ById(id string) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doById(CREDENTIAL_TYPE, id, resp)
	return resp, err
}

func (c *CredentialClient) Delete(container *Credential) error {
	return c.rancherClient.doResourceDelete(CREDENTIAL_TYPE, &container.Resource)
}

func (c *CredentialClient) ActionActivate(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *CredentialClient) ActionCreate(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *CredentialClient) ActionDeactivate(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *CredentialClient) ActionPurge(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *CredentialClient) ActionRemove(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *CredentialClient) ActionRestore(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *CredentialClient) ActionUpdate(resource *Credential) (*Credential, error) {
	resp := &Credential{}
	err := c.rancherClient.doEmptyAction(CREDENTIAL_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
