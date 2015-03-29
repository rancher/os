package client

const (
	PROCESS_EXECUTION_TYPE = "processExecution"
)

type ProcessExecution struct {
	Resource
    
    Log map[string]interface{} `json:"log,omitempty"`
    
    ProcessInstanceId string `json:"processInstanceId,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type ProcessExecutionCollection struct {
	Collection
	Data []ProcessExecution `json:"data,omitempty"`
}

type ProcessExecutionClient struct {
	rancherClient *RancherClient
}

type ProcessExecutionOperations interface {
	List(opts *ListOpts) (*ProcessExecutionCollection, error)
	Create(opts *ProcessExecution) (*ProcessExecution, error)
	Update(existing *ProcessExecution, updates interface{}) (*ProcessExecution, error)
	ById(id string) (*ProcessExecution, error)
	Delete(container *ProcessExecution) error
}

func newProcessExecutionClient(rancherClient *RancherClient) *ProcessExecutionClient {
	return &ProcessExecutionClient{
		rancherClient: rancherClient,
	}
}

func (c *ProcessExecutionClient) Create(container *ProcessExecution) (*ProcessExecution, error) {
	resp := &ProcessExecution{}
	err := c.rancherClient.doCreate(PROCESS_EXECUTION_TYPE, container, resp)
	return resp, err
}

func (c *ProcessExecutionClient) Update(existing *ProcessExecution, updates interface{}) (*ProcessExecution, error) {
	resp := &ProcessExecution{}
	err := c.rancherClient.doUpdate(PROCESS_EXECUTION_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ProcessExecutionClient) List(opts *ListOpts) (*ProcessExecutionCollection, error) {
	resp := &ProcessExecutionCollection{}
	err := c.rancherClient.doList(PROCESS_EXECUTION_TYPE, opts, resp)
	return resp, err
}

func (c *ProcessExecutionClient) ById(id string) (*ProcessExecution, error) {
	resp := &ProcessExecution{}
	err := c.rancherClient.doById(PROCESS_EXECUTION_TYPE, id, resp)
	return resp, err
}

func (c *ProcessExecutionClient) Delete(container *ProcessExecution) error {
	return c.rancherClient.doResourceDelete(PROCESS_EXECUTION_TYPE, &container.Resource)
}
