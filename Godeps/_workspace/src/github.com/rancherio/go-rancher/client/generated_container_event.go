package client

const (
	CONTAINER_EVENT_TYPE = "containerEvent"
)

type ContainerEvent struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    DockerInspect interface{} `json:"dockerInspect,omitempty"`
    
    ExternalFrom string `json:"externalFrom,omitempty"`
    
    ExternalId string `json:"externalId,omitempty"`
    
    ExternalStatus string `json:"externalStatus,omitempty"`
    
    ExternalTimestamp int `json:"externalTimestamp,omitempty"`
    
    HostId string `json:"hostId,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    ReportedHostUuid string `json:"reportedHostUuid,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
}

type ContainerEventCollection struct {
	Collection
	Data []ContainerEvent `json:"data,omitempty"`
}

type ContainerEventClient struct {
	rancherClient *RancherClient
}

type ContainerEventOperations interface {
	List(opts *ListOpts) (*ContainerEventCollection, error)
	Create(opts *ContainerEvent) (*ContainerEvent, error)
	Update(existing *ContainerEvent, updates interface{}) (*ContainerEvent, error)
	ById(id string) (*ContainerEvent, error)
	Delete(container *ContainerEvent) error
    ActionCreate (*ContainerEvent) (*ContainerEvent, error)
    ActionRemove (*ContainerEvent) (*ContainerEvent, error)
}

func newContainerEventClient(rancherClient *RancherClient) *ContainerEventClient {
	return &ContainerEventClient{
		rancherClient: rancherClient,
	}
}

func (c *ContainerEventClient) Create(container *ContainerEvent) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doCreate(CONTAINER_EVENT_TYPE, container, resp)
	return resp, err
}

func (c *ContainerEventClient) Update(existing *ContainerEvent, updates interface{}) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doUpdate(CONTAINER_EVENT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ContainerEventClient) List(opts *ListOpts) (*ContainerEventCollection, error) {
	resp := &ContainerEventCollection{}
	err := c.rancherClient.doList(CONTAINER_EVENT_TYPE, opts, resp)
	return resp, err
}

func (c *ContainerEventClient) ById(id string) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doById(CONTAINER_EVENT_TYPE, id, resp)
	return resp, err
}

func (c *ContainerEventClient) Delete(container *ContainerEvent) error {
	return c.rancherClient.doResourceDelete(CONTAINER_EVENT_TYPE, &container.Resource)
}

func (c *ContainerEventClient) ActionCreate(resource *ContainerEvent) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doEmptyAction(CONTAINER_EVENT_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *ContainerEventClient) ActionRemove(resource *ContainerEvent) (*ContainerEvent, error) {
	resp := &ContainerEvent{}
	err := c.rancherClient.doEmptyAction(CONTAINER_EVENT_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}
