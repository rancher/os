package client

const (
	PHYSICAL_HOST_TYPE = "physicalHost"
)

type PhysicalHost struct {
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

type PhysicalHostCollection struct {
	Collection
	Data []PhysicalHost `json:"data,omitempty"`
}

type PhysicalHostClient struct {
	rancherClient *RancherClient
}

type PhysicalHostOperations interface {
	List(opts *ListOpts) (*PhysicalHostCollection, error)
	Create(opts *PhysicalHost) (*PhysicalHost, error)
	Update(existing *PhysicalHost, updates interface{}) (*PhysicalHost, error)
	ById(id string) (*PhysicalHost, error)
	Delete(container *PhysicalHost) error
    ActionBootstrap (*PhysicalHost) (*PhysicalHost, error)
    ActionCreate (*PhysicalHost) (*PhysicalHost, error)
    ActionRemove (*PhysicalHost) (*PhysicalHost, error)
    ActionUpdate (*PhysicalHost) (*PhysicalHost, error)
}

func newPhysicalHostClient(rancherClient *RancherClient) *PhysicalHostClient {
	return &PhysicalHostClient{
		rancherClient: rancherClient,
	}
}

func (c *PhysicalHostClient) Create(container *PhysicalHost) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doCreate(PHYSICAL_HOST_TYPE, container, resp)
	return resp, err
}

func (c *PhysicalHostClient) Update(existing *PhysicalHost, updates interface{}) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doUpdate(PHYSICAL_HOST_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *PhysicalHostClient) List(opts *ListOpts) (*PhysicalHostCollection, error) {
	resp := &PhysicalHostCollection{}
	err := c.rancherClient.doList(PHYSICAL_HOST_TYPE, opts, resp)
	return resp, err
}

func (c *PhysicalHostClient) ById(id string) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doById(PHYSICAL_HOST_TYPE, id, resp)
	return resp, err
}

func (c *PhysicalHostClient) Delete(container *PhysicalHost) error {
	return c.rancherClient.doResourceDelete(PHYSICAL_HOST_TYPE, &container.Resource)
}

func (c *PhysicalHostClient) ActionBootstrap(resource *PhysicalHost) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doEmptyAction(PHYSICAL_HOST_TYPE, "bootstrap", &resource.Resource, resp)
	return resp, err
}

func (c *PhysicalHostClient) ActionCreate(resource *PhysicalHost) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doEmptyAction(PHYSICAL_HOST_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *PhysicalHostClient) ActionRemove(resource *PhysicalHost) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doEmptyAction(PHYSICAL_HOST_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *PhysicalHostClient) ActionUpdate(resource *PhysicalHost) (*PhysicalHost, error) {
	resp := &PhysicalHost{}
	err := c.rancherClient.doEmptyAction(PHYSICAL_HOST_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
