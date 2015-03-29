package client

const (
	IMAGE_TYPE = "image"
)

type Image struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
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

type ImageCollection struct {
	Collection
	Data []Image `json:"data,omitempty"`
}

type ImageClient struct {
	rancherClient *RancherClient
}

type ImageOperations interface {
	List(opts *ListOpts) (*ImageCollection, error)
	Create(opts *Image) (*Image, error)
	Update(existing *Image, updates interface{}) (*Image, error)
	ById(id string) (*Image, error)
	Delete(container *Image) error
    ActionActivate (*Image) (*Image, error)
    ActionCreate (*Image) (*Image, error)
    ActionDeactivate (*Image) (*Image, error)
    ActionPurge (*Image) (*Image, error)
    ActionRemove (*Image) (*Image, error)
    ActionRestore (*Image) (*Image, error)
    ActionUpdate (*Image) (*Image, error)
}

func newImageClient(rancherClient *RancherClient) *ImageClient {
	return &ImageClient{
		rancherClient: rancherClient,
	}
}

func (c *ImageClient) Create(container *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doCreate(IMAGE_TYPE, container, resp)
	return resp, err
}

func (c *ImageClient) Update(existing *Image, updates interface{}) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doUpdate(IMAGE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ImageClient) List(opts *ListOpts) (*ImageCollection, error) {
	resp := &ImageCollection{}
	err := c.rancherClient.doList(IMAGE_TYPE, opts, resp)
	return resp, err
}

func (c *ImageClient) ById(id string) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doById(IMAGE_TYPE, id, resp)
	return resp, err
}

func (c *ImageClient) Delete(container *Image) error {
	return c.rancherClient.doResourceDelete(IMAGE_TYPE, &container.Resource)
}

func (c *ImageClient) ActionActivate(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "activate", &resource.Resource, resp)
	return resp, err
}

func (c *ImageClient) ActionCreate(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "create", &resource.Resource, resp)
	return resp, err
}

func (c *ImageClient) ActionDeactivate(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "deactivate", &resource.Resource, resp)
	return resp, err
}

func (c *ImageClient) ActionPurge(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "purge", &resource.Resource, resp)
	return resp, err
}

func (c *ImageClient) ActionRemove(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "remove", &resource.Resource, resp)
	return resp, err
}

func (c *ImageClient) ActionRestore(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "restore", &resource.Resource, resp)
	return resp, err
}

func (c *ImageClient) ActionUpdate(resource *Image) (*Image, error) {
	resp := &Image{}
	err := c.rancherClient.doEmptyAction(IMAGE_TYPE, "update", &resource.Resource, resp)
	return resp, err
}
