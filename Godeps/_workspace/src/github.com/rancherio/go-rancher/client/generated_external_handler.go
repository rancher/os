package client

const (
	EXTERNAL_HANDLER_TYPE = "externalHandler"
)

type ExternalHandler struct {
	Resource
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    Priority int `json:"priority,omitempty"`
    
    ProcessNames []string `json:"processNames,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    Retries int `json:"retries,omitempty"`
    
    State string `json:"state,omitempty"`
    
    TimeoutMillis int `json:"timeoutMillis,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type ExternalHandlerCollection struct {
	Collection
	Data []ExternalHandler `json:"data,omitempty"`
}

type ExternalHandlerClient struct {
	rancherClient *RancherClient
}

type ExternalHandlerOperations interface {
	List(opts *ListOpts) (*ExternalHandlerCollection, error)
	Create(opts *ExternalHandler) (*ExternalHandler, error)
	Update(existing *ExternalHandler, updates interface{}) (*ExternalHandler, error)
	ById(id string) (*ExternalHandler, error)
	Delete(container *ExternalHandler) error
    ActionActivate (*ExternalHandler) (*ExternalHandler, error)
    ActionCreate (*ExternalHandler) (*ExternalHandler, error)
    ActionDeactivate (*ExternalHandler) (*ExternalHandler, error)
    ActionPurge (*ExternalHandler) (*ExternalHandler, error)
    ActionRemove (*ExternalHandler) (*ExternalHandler, error)
    ActionRestore (*ExternalHandler) (*ExternalHandler, error)
    ActionUpdate (*ExternalHandler) (*ExternalHandler, error)
}

func newExternalHandlerClient(rancherClient *RancherClient) *ExternalHandlerClient {
	return &ExternalHandlerClient{
		rancherClient: rancherClient,
	}
}

func (c *ExternalHandlerClient) Create(container *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doCreate(EXTERNAL_HANDLER_TYPE, container, resp)
	return resp, err
}

func (c *ExternalHandlerClient) Update(existing *ExternalHandler, updates interface{}) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doUpdate(EXTERNAL_HANDLER_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExternalHandlerClient) List(opts *ListOpts) (*ExternalHandlerCollection, error) {
	resp := &ExternalHandlerCollection{}
	err := c.rancherClient.doList(EXTERNAL_HANDLER_TYPE, opts, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ById(id string) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doById(EXTERNAL_HANDLER_TYPE, id, resp)
	return resp, err
}

func (c *ExternalHandlerClient) Delete(container *ExternalHandler) error {
	return c.rancherClient.doResourceDelete(EXTERNAL_HANDLER_TYPE, &container.Resource)
}

func (c *ExternalHandlerClient) ActionActivate(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ActionCreate(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ActionDeactivate(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ActionPurge(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ActionRemove(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ActionRestore(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerClient) ActionUpdate(resource *ExternalHandler) (*ExternalHandler, error) {
	resp := &ExternalHandler{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
