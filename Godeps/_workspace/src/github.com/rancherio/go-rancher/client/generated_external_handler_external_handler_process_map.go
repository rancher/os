package client

const (
	EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE = "externalHandlerExternalHandlerProcessMap"
)

type ExternalHandlerExternalHandlerProcessMap struct {
	Resource
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    ExternalHandlerId string `json:"externalHandlerId,omitempty"`
    
    ExternalHandlerProcessId string `json:"externalHandlerProcessId,omitempty"`
    
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

type ExternalHandlerExternalHandlerProcessMapCollection struct {
	Collection
	Data []ExternalHandlerExternalHandlerProcessMap `json:"data,omitempty"`
}

type ExternalHandlerExternalHandlerProcessMapClient struct {
	rancherClient *RancherClient
}

type ExternalHandlerExternalHandlerProcessMapOperations interface {
	List(opts *ListOpts) (*ExternalHandlerExternalHandlerProcessMapCollection, error)
	Create(opts *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
	Update(existing *ExternalHandlerExternalHandlerProcessMap, updates interface{}) (*ExternalHandlerExternalHandlerProcessMap, error)
	ById(id string) (*ExternalHandlerExternalHandlerProcessMap, error)
	Delete(container *ExternalHandlerExternalHandlerProcessMap) error
    ActionActivate (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
    ActionCreate (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
    ActionDeactivate (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
    ActionPurge (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
    ActionRemove (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
    ActionRestore (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
    ActionUpdate (*ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error)
}

func newExternalHandlerExternalHandlerProcessMapClient(rancherClient *RancherClient) *ExternalHandlerExternalHandlerProcessMapClient {
	return &ExternalHandlerExternalHandlerProcessMapClient{
		rancherClient: rancherClient,
	}
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) Create(container *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doCreate(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, container, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) Update(existing *ExternalHandlerExternalHandlerProcessMap, updates interface{}) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doUpdate(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) List(opts *ListOpts) (*ExternalHandlerExternalHandlerProcessMapCollection, error) {
	resp := &ExternalHandlerExternalHandlerProcessMapCollection{}
	err := c.rancherClient.doList(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, opts, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ById(id string) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doById(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, id, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) Delete(container *ExternalHandlerExternalHandlerProcessMap) error {
	return c.rancherClient.doResourceDelete(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, &container.Resource)
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionActivate(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionCreate(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionDeactivate(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionPurge(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionRemove(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionRestore(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *ExternalHandlerExternalHandlerProcessMapClient) ActionUpdate(resource *ExternalHandlerExternalHandlerProcessMap) (*ExternalHandlerExternalHandlerProcessMap, error) {
	resp := &ExternalHandlerExternalHandlerProcessMap{}
	err := c.rancherClient.doEmptyAction(EXTERNAL_HANDLER_EXTERNAL_HANDLER_PROCESS_MAP_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
