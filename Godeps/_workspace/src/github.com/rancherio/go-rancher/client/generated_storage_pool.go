package client

const (
	STORAGE_POOL_TYPE = "storagePool"
)

type StoragePool struct {
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

type StoragePoolCollection struct {
	Collection
	Data []StoragePool `json:"data,omitempty"`
}

type StoragePoolClient struct {
	rancherClient *RancherClient
}

type StoragePoolOperations interface {
	List(opts *ListOpts) (*StoragePoolCollection, error)
	Create(opts *StoragePool) (*StoragePool, error)
	Update(existing *StoragePool, updates interface{}) (*StoragePool, error)
	ById(id string) (*StoragePool, error)
	Delete(container *StoragePool) error
    ActionActivate (*StoragePool) (*StoragePool, error)
    ActionCreate (*StoragePool) (*StoragePool, error)
    ActionDeactivate (*StoragePool) (*StoragePool, error)
    ActionPurge (*StoragePool) (*StoragePool, error)
    ActionRemove (*StoragePool) (*StoragePool, error)
    ActionRestore (*StoragePool) (*StoragePool, error)
    ActionUpdate (*StoragePool) (*StoragePool, error)
}

func newStoragePoolClient(rancherClient *RancherClient) *StoragePoolClient {
	return &StoragePoolClient{
		rancherClient: rancherClient,
	}
}

func (c *StoragePoolClient) Create(container *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doCreate(STORAGE_POOL_TYPE, container, resp)
	return resp, err
}

func (c *StoragePoolClient) Update(existing *StoragePool, updates interface{}) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doUpdate(STORAGE_POOL_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *StoragePoolClient) List(opts *ListOpts) (*StoragePoolCollection, error) {
	resp := &StoragePoolCollection{}
	err := c.rancherClient.doList(STORAGE_POOL_TYPE, opts, resp)
	return resp, err
}

func (c *StoragePoolClient) ById(id string) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doById(STORAGE_POOL_TYPE, id, resp)
	return resp, err
}

func (c *StoragePoolClient) Delete(container *StoragePool) error {
	return c.rancherClient.doResourceDelete(STORAGE_POOL_TYPE, &container.Resource)
}

func (c *StoragePoolClient) ActionActivate(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *StoragePoolClient) ActionCreate(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *StoragePoolClient) ActionDeactivate(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *StoragePoolClient) ActionPurge(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *StoragePoolClient) ActionRemove(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *StoragePoolClient) ActionRestore(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *StoragePoolClient) ActionUpdate(resource *StoragePool) (*StoragePool, error) {
	resp := &StoragePool{}
	err := c.rancherClient.doEmptyAction(STORAGE_POOL_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
