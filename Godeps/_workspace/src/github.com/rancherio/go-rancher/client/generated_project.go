package client

const (
	PROJECT_TYPE = "project"
)

type Project struct {
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

type ProjectCollection struct {
	Collection
	Data []Project `json:"data,omitempty"`
}

type ProjectClient struct {
	rancherClient *RancherClient
}

type ProjectOperations interface {
	List(opts *ListOpts) (*ProjectCollection, error)
	Create(opts *Project) (*Project, error)
	Update(existing *Project, updates interface{}) (*Project, error)
	ById(id string) (*Project, error)
	Delete(container *Project) error
}

func newProjectClient(rancherClient *RancherClient) *ProjectClient {
	return &ProjectClient{
		rancherClient: rancherClient,
	}
}

func (c *ProjectClient) Create(container *Project) (*Project, error) {
	resp := &Project{}
	err := c.rancherClient.doCreate(PROJECT_TYPE, container, resp)
	return resp, err
}

func (c *ProjectClient) Update(existing *Project, updates interface{}) (*Project, error) {
	resp := &Project{}
	err := c.rancherClient.doUpdate(PROJECT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProjectClient) List(opts *ListOpts) (*ProjectCollection, error) {
	resp := &ProjectCollection{}
	err := c.rancherClient.doList(PROJECT_TYPE, opts, resp)
	return resp, err
}

func (c *ProjectClient) ById(id string) (*Project, error) {
	resp := &Project{}
	err := c.rancherClient.doById(PROJECT_TYPE, id, resp)
	return resp, err
}

func (c *ProjectClient) Delete(container *Project) error {
	return c.rancherClient.doResourceDelete(PROJECT_TYPE, &container.Resource)
}
