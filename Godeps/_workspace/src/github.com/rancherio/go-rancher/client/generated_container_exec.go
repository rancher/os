package client

const (
	CONTAINER_EXEC_TYPE = "containerExec"
)

type ContainerExec struct {
	Resource
    
    AttachStdin bool `json:"attachStdin,omitempty"`
    
    AttachStdout bool `json:"attachStdout,omitempty"`
    
    Command []string `json:"command,omitempty"`
    
    Tty bool `json:"tty,omitempty"`
    
}

type ContainerExecCollection struct {
	Collection
	Data []ContainerExec `json:"data,omitempty"`
}

type ContainerExecClient struct {
	rancherClient *RancherClient
}

type ContainerExecOperations interface {
	List(opts *ListOpts) (*ContainerExecCollection, error)
	Create(opts *ContainerExec) (*ContainerExec, error)
	Update(existing *ContainerExec, updates interface{}) (*ContainerExec, error)
	ById(id string) (*ContainerExec, error)
	Delete(container *ContainerExec) error
}

func newContainerExecClient(rancherClient *RancherClient) *ContainerExecClient {
	return &ContainerExecClient{
		rancherClient: rancherClient,
	}
}

func (c *ContainerExecClient) Create(container *ContainerExec) (*ContainerExec, error) {
	resp := &ContainerExec{}
	err := c.rancherClient.doCreate(CONTAINER_EXEC_TYPE, container, resp)
	return resp, err
}

func (c *ContainerExecClient) Update(existing *ContainerExec, updates interface{}) (*ContainerExec, error) {
	resp := &ContainerExec{}
	err := c.rancherClient.doUpdate(CONTAINER_EXEC_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ContainerExecClient) List(opts *ListOpts) (*ContainerExecCollection, error) {
	resp := &ContainerExecCollection{}
	err := c.rancherClient.doList(CONTAINER_EXEC_TYPE, opts, resp)
	return resp, err
}

func (c *ContainerExecClient) ById(id string) (*ContainerExec, error) {
	resp := &ContainerExec{}
	err := c.rancherClient.doById(CONTAINER_EXEC_TYPE, id, resp)
	return resp, err
}

func (c *ContainerExecClient) Delete(container *ContainerExec) error {
	return c.rancherClient.doResourceDelete(CONTAINER_EXEC_TYPE, &container.Resource)
}
