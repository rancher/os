package client

const (
	CERTIFICATE_TYPE = "certificate"
)

type Certificate struct {
	Resource
    
    AccountId string `json:"accountId,omitempty"`
    
    Cert string `json:"cert,omitempty"`
    
    CertChain string `json:"certChain,omitempty"`
    
    Created string `json:"created,omitempty"`
    
    Data map[string]interface{} `json:"data,omitempty"`
    
    Description string `json:"description,omitempty"`
    
    Key string `json:"key,omitempty"`
    
    Kind string `json:"kind,omitempty"`
    
    Name string `json:"name,omitempty"`
    
    RemoveTime string `json:"removeTime,omitempty"`
    
    Removed string `json:"removed,omitempty"`
    
    State string `json:"state,omitempty"`
    
    Uuid string `json:"uuid,omitempty"`
    
}

type CertificateCollection struct {
	Collection
	Data []Certificate `json:"data,omitempty"`
}

type CertificateClient struct {
	rancherClient *RancherClient
}

type CertificateOperations interface {
	List(opts *ListOpts) (*CertificateCollection, error)
	Create(opts *Certificate) (*Certificate, error)
	Update(existing *Certificate, updates interface{}) (*Certificate, error)
	ById(id string) (*Certificate, error)
	Delete(container *Certificate) error
}

func newCertificateClient(rancherClient *RancherClient) *CertificateClient {
	return &CertificateClient{
		rancherClient: rancherClient,
	}
}

func (c *CertificateClient) Create(container *Certificate) (*Certificate, error) {
	resp := &Certificate{}
	err := c.rancherClient.doCreate(CERTIFICATE_TYPE, container, resp)
	return resp, err
}

func (c *CertificateClient) Update(existing *Certificate, updates interface{}) (*Certificate, error) {
	resp := &Certificate{}
	err := c.rancherClient.doUpdate(CERTIFICATE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *CertificateClient) List(opts *ListOpts) (*CertificateCollection, error) {
	resp := &CertificateCollection{}
	err := c.rancherClient.doList(CERTIFICATE_TYPE, opts, resp)
	return resp, err
}

func (c *CertificateClient) ById(id string) (*Certificate, error) {
	resp := &Certificate{}
	err := c.rancherClient.doById(CERTIFICATE_TYPE, id, resp)
	return resp, err
}

func (c *CertificateClient) Delete(container *Certificate) error {
	return c.rancherClient.doResourceDelete(CERTIFICATE_TYPE, &container.Resource)
}
