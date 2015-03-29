package client

const (
	REGISTRY_CREDENTIAL_TYPE = "registryCredential"
)

type RegistryCredential struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Email string `json:"email,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    PublicValue string `json:"publicValue,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    SecretValue string `json:"secretValue,omitempty"`
    
    State string `json:"state,omitempty"`
    
    StoragePoolId string `json:"storagePoolId,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type RegistryCredentialCollection struct {
	Collection
	Data []RegistryCredential `json:"data,omitempty"`
}

type RegistryCredentialClient struct {
	rancherClient *RancherClient
}

type RegistryCredentialOperations interface {
	List(opts *ListOpts) (*RegistryCredentialCollection, error)
	Create(opts *RegistryCredential) (*RegistryCredential, error)
	Update(existing *RegistryCredential, updates interface{}) (*RegistryCredential, error)
	ById(id string) (*RegistryCredential, error)
	Delete(container *RegistryCredential) error
}

func newRegistryCredentialClient(rancherClient *RancherClient) *RegistryCredentialClient {
	return &RegistryCredentialClient{
		rancherClient: rancherClient,
	}
}

func (c *RegistryCredentialClient) Create(container *RegistryCredential) (*RegistryCredential, error) {
	resp := &RegistryCredential{}
	err := c.rancherClient.doCreate(REGISTRY_CREDENTIAL_TYPE, container, resp)
	return resp, err
}

func (c *RegistryCredentialClient) Update(existing *RegistryCredential, updates interface{}) (*RegistryCredential, error) {
	resp := &RegistryCredential{}
	err := c.rancherClient.doUpdate(REGISTRY_CREDENTIAL_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *RegistryCredentialClient) List(opts *ListOpts) (*RegistryCredentialCollection, error) {
	resp := &RegistryCredentialCollection{}
	err := c.rancherClient.doList(REGISTRY_CREDENTIAL_TYPE, opts, resp)
	return resp, err
}

func (c *RegistryCredentialClient) ById(id string) (*RegistryCredential, error) {
	resp := &RegistryCredential{}
	err := c.rancherClient.doById(REGISTRY_CREDENTIAL_TYPE, id, resp)
	return resp, err
}

func (c *RegistryCredentialClient) Delete(container *RegistryCredential) error {
	return c.rancherClient.doResourceDelete(REGISTRY_CREDENTIAL_TYPE, &container.Resource)
}
