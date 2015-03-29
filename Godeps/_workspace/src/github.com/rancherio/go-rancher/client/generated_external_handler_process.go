package client

const (
	EXTERNAL_HANDLER_PROCESS_TYPE = "externalHandlerProcess"
)

type ExternalHandlerProcess struct {
	Resource
    
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

type ExternalHandlerProcessCollection struct {
	Collection
	Data []ExternalHandlerProcess `json:"data,omitempty"`
}

type ExternalHandlerProcessClient struct {
	rancherClient *RancherClient
}

type ExternalHandlerProcessOperations interface {
	List(opts *ListOpts) (*ExternalHandlerProcessCollection, error)
	Create(opts *ExternalHandlerProcess) (*ExternalHandlerProcess, error)
	Update(existing *ExternalHandlerProcess, updates interface{}) (*ExternalHandlerProcess, error)
	ById(id string) (*ExternalHandlerProcess, error)
	Delete(container *ExternalHandlerProcess) error
    ActionActivate (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
    ActionCreate (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
    ActionDeactivate (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
    ActionPurge (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
    ActionRemove (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
    ActionRestore (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
    ActionUpdate (*ExternalHandlerProcess) (*ExternalHandlerProcess, error)
}

func newExternalHandlerProcessClient(rancherClient *RancherClient) *ExternalHandlerProcessClient {
	return &ExternalHandlerProcessClient{
		rancherClient: rancherClient,
	}
}

func (c *ExternalHandlerProcessClient) Create(container *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doCreate(EXTERNAL_HANDLER_PROCESS_TYPE, container, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) Update(existing *ExternalHandlerProcess, updates interface{}) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doUpdate(EXTERNAL_HANDLER_PROCESS_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) List(opts *ListOpts) (*ExternalHandlerProcessCollection, error) {
	resp := &ExternalHandlerProcessCollection{}
	err := c.rancherClient.doList(EXTERNAL_HANDLER_PROCESS_TYPE, opts, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ById(id string) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doById(EXTERNAL_HANDLER_PROCESS_TYPE, id, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) Delete(container *ExternalHandlerProcess) error {
	return c.rancherClient.doResourceDelete(EXTERNAL_HANDLER_PROCESS_TYPE, &container.Resource)
}

func (c *ExternalHandlerProcessClient) ActionActivate(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ActionCreate(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ActionDeactivate(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ActionPurge(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ActionRemove(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ActionRestore(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerProcessClient) ActionUpdate(resource *ExternalHandlerProcess) (*ExternalHandlerProcess, error) {
	resp := &ExternalHandlerProcess{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_PROCESS_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
