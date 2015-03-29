package client

const (
	AGENT_TYPE = "agent"
)

type Agent struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    ManagedConfig bool `json:"managedConfig,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Transitioning string `json:"transitioning,omitempty"`
    
    TransitioningMessage string `json:"transitioningMessage,omitempty"`
    
    TransitioningProgress int `json:"transitioningProgress,omitempty"`
    
    Uri string `json:"uri,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type AgentCollection struct {
	Collection
	Data []Agent `json:"data,omitempty"`
}

type AgentClient struct {
	rancherClient *RancherClient
}

type AgentOperations interface {
	List(opts *ListOpts) (*AgentCollection, error)
	Create(opts *Agent) (*Agent, error)
	Update(existing *Agent, updates interface{}) (*Agent, error)
	ById(id string) (*Agent, error)
	Delete(container *Agent) error
    ActionActivate (*Agent) (*Agent, error)
    ActionCreate (*Agent) (*Agent, error)
    ActionDeactivate (*Agent) (*Agent, error)
    ActionPurge (*Agent) (*Agent, error)
    ActionReconnect (*Agent) (*Agent, error)
    ActionRemove (*Agent) (*Agent, error)
    ActionRestore (*Agent) (*Agent, error)
    ActionUpdate (*Agent) (*Agent, error)
}

func newAgentClient(rancherClient *RancherClient) *AgentClient {
	return &AgentClient{
		rancherClient: rancherClient,
	}
}

func (c *AgentClient) Create(container *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doCreate(AGENT_TYPE, container, resp)
	return resp, err
}

func (c *AgentClient) Update(existing *Agent, updates interface{}) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doUpdate(AGENT_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *AgentClient) List(opts *ListOpts) (*AgentCollection, error) {
	resp := &AgentCollection{}
	err := c.rancherClient.doList(AGENT_TYPE, opts, resp)
	return resp, err
}

func (c *AgentClient) ById(id string) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doById(AGENT_TYPE, id, resp)
	return resp, err
}

func (c *AgentClient) Delete(container *Agent) error {
	return c.rancherClient.doResourceDelete(AGENT_TYPE, &container.Resource)
}

func (c *AgentClient) ActionActivate(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionCreate(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionDeactivate(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionPurge(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionReconnect(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "reconnect", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionRemove(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionRestore(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *AgentClient) ActionUpdate(resource *Agent) (*Agent, error) {
	resp := &Agent{}
	err := c.rancherClient.doEmptyAction(AGENT_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
