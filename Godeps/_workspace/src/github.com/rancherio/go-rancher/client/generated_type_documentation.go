package client

const (
	TYPE_DOCUMENTATION_TYPE = "typeDocumentation"
)

type TypeDocumentation struct {
	Resource
    
    Description string `json:"description,omitempty"`
    
}

type TypeDocumentationCollection struct {
	Collection
	Data []TypeDocumentation `json:"data,omitempty"`
}

type TypeDocumentationClient struct {
	rancherClient *RancherClient
}

type TypeDocumentationOperations interface {
	List(opts *ListOpts) (*TypeDocumentationCollection, error)
	Create(opts *TypeDocumentation) (*TypeDocumentation, error)
	Update(existing *TypeDocumentation, updates interface{}) (*TypeDocumentation, error)
	ById(id string) (*TypeDocumentation, error)
	Delete(container *TypeDocumentation) error
}

func newTypeDocumentationClient(rancherClient *RancherClient) *TypeDocumentationClient {
	return &TypeDocumentationClient{
		rancherClient: rancherClient,
	}
}

func (c *TypeDocumentationClient) Create(container *TypeDocumentation) (*TypeDocumentation, error) {
	resp := &TypeDocumentation{}
	err := c.rancherClient.doCreate(TYPE_DOCUMENTATION_TYPE, container, resp)
	return resp, err
}

func (c *TypeDocumentationClient) Update(existing *TypeDocumentation, updates interface{}) (*TypeDocumentation, error) {
	resp := &TypeDocumentation{}
	err := c.rancherClient.doUpdate(TYPE_DOCUMENTATION_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *TypeDocumentationClient) List(opts *ListOpts) (*TypeDocumentationCollection, error) {
	resp := &TypeDocumentationCollection{}
	err := c.rancherClient.doList(TYPE_DOCUMENTATION_TYPE, opts, resp)
	return resp, err
}

func (c *TypeDocumentationClient) ById(id string) (*TypeDocumentation, error) {
	resp := &TypeDocumentation{}
	err := c.rancherClient.doById(TYPE_DOCUMENTATION_TYPE, id, resp)
	return resp, err
}

func (c *TypeDocumentationClient) Delete(container *TypeDocumentation) error {
	return c.rancherClient.doResourceDelete(TYPE_DOCUMENTATION_TYPE, &container.Resource)
}
