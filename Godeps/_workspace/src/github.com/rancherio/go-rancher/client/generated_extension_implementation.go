package client

const (
	EXTENSION_IMPLEMENTATION_TYPE = "extensionImplementation"
)

type ExtensionImplementation struct {
	Resource
    
    ClassName string `json:"className,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    Properties map[string]interface{} `json:"properties,omitempty"`
    
}

type ExtensionImplementationCollection struct {
	Collection
	Data []ExtensionImplementation `json:"data,omitempty"`
}

type ExtensionImplementationClient struct {
	rancherClient *RancherClient
}

type ExtensionImplementationOperations interface {
	List(opts *ListOpts) (*ExtensionImplementationCollection, error)
	Create(opts *ExtensionImplementation) (*ExtensionImplementation, error)
	Update(existing *ExtensionImplementation, updates interface{}) (*ExtensionImplementation, error)
	ById(id string) (*ExtensionImplementation, error)
	Delete(container *ExtensionImplementation) error
}

func newExtensionImplementationClient(rancherClient *RancherClient) *ExtensionImplementationClient {
	return &ExtensionImplementationClient{
		rancherClient: rancherClient,
	}
}

func (c *ExtensionImplementationClient) Create(container *ExtensionImplementation) (*ExtensionImplementation, error) {
	resp := &ExtensionImplementation{}
	err := c.rancherClient.doCreate(EXTENSION_IMPLEMENTATION_TYPE, container, resp)
	return resp, err
}

func (c *ExtensionImplementationClient) Update(existing *ExtensionImplementation, updates interface{}) (*ExtensionImplementation, error) {
	resp := &ExtensionImplementation{}
	err := c.rancherClient.doUpdate(EXTENSION_IMPLEMENTATION_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ExtensionImplementationClient) List(opts *ListOpts) (*ExtensionImplementationCollection, error) {
	resp := &ExtensionImplementationCollection{}
	err := c.rancherClient.doList(EXTENSION_IMPLEMENTATION_TYPE, opts, resp)
	return resp, err
}

func (c *ExtensionImplementationClient) ById(id string) (*ExtensionImplementation, error) {
	resp := &ExtensionImplementation{}
	err := c.rancherClient.doById(EXTENSION_IMPLEMENTATION_TYPE, id, resp)
	return resp, err
}

func (c *ExtensionImplementationClient) Delete(container *ExtensionImplementation) error {
	return c.rancherClient.doResourceDelete(EXTENSION_IMPLEMENTATION_TYPE, &container.Resource)
}
