package client

const (
	COMPOSE_SERVICE_TYPE = "composeService"
)

type ComposeService struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	CurrentScale int64 `json:"currentScale,omitempty" yaml:"current_scale,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	Fqdn string `json:"fqdn,omitempty" yaml:"fqdn,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	InstanceIds []string `json:"instanceIds,omitempty" yaml:"instance_ids,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	LaunchConfig *LaunchConfig `json:"launchConfig,omitempty" yaml:"launch_config,omitempty"`

	LinkedServices map[string]interface{} `json:"linkedServices,omitempty" yaml:"linked_services,omitempty"`

	Name string `json:"name,omitempty" yaml:"name,omitempty"`

	PublicEndpoints []PublicEndpoint `json:"publicEndpoints,omitempty" yaml:"public_endpoints,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	Scale int64 `json:"scale,omitempty" yaml:"scale,omitempty"`

	ScalePolicy *ScalePolicy `json:"scalePolicy,omitempty" yaml:"scale_policy,omitempty"`

	SelectorContainer string `json:"selectorContainer,omitempty" yaml:"selector_container,omitempty"`

	SelectorLink string `json:"selectorLink,omitempty" yaml:"selector_link,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartOnCreate bool `json:"startOnCreate,omitempty" yaml:"start_on_create,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	System bool `json:"system,omitempty" yaml:"system,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vip string `json:"vip,omitempty" yaml:"vip,omitempty"`
}

type ComposeServiceCollection struct {
	Collection
	Data   []ComposeService `json:"data,omitempty"`
	client *ComposeServiceClient
}

type ComposeServiceClient struct {
	rancherClient *RancherClient
}

type ComposeServiceOperations interface {
	List(opts *ListOpts) (*ComposeServiceCollection, error)
	Create(opts *ComposeService) (*ComposeService, error)
	Update(existing *ComposeService, updates interface{}) (*ComposeService, error)
	ById(id string) (*ComposeService, error)
	Delete(container *ComposeService) error

	ActionActivate(*ComposeService) (*Service, error)

	ActionCancelupgrade(*ComposeService) (*Service, error)

	ActionContinueupgrade(*ComposeService) (*Service, error)

	ActionCreate(*ComposeService) (*Service, error)

	ActionFinishupgrade(*ComposeService) (*Service, error)

	ActionRemove(*ComposeService) (*Service, error)

	ActionRollback(*ComposeService) (*Service, error)
}

func newComposeServiceClient(rancherClient *RancherClient) *ComposeServiceClient {
	return &ComposeServiceClient{
		rancherClient: rancherClient,
	}
}

func (c *ComposeServiceClient) Create(container *ComposeService) (*ComposeService, error) {
	resp := &ComposeService{}
	err := c.rancherClient.doCreate(COMPOSE_SERVICE_TYPE, container, resp)
	return resp, err
}

func (c *ComposeServiceClient) Update(existing *ComposeService, updates interface{}) (*ComposeService, error) {
	resp := &ComposeService{}
	err := c.rancherClient.doUpdate(COMPOSE_SERVICE_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *ComposeServiceClient) List(opts *ListOpts) (*ComposeServiceCollection, error) {
	resp := &ComposeServiceCollection{}
	err := c.rancherClient.doList(COMPOSE_SERVICE_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *ComposeServiceCollection) Next() (*ComposeServiceCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &ComposeServiceCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *ComposeServiceClient) ById(id string) (*ComposeService, error) {
	resp := &ComposeService{}
	err := c.rancherClient.doById(COMPOSE_SERVICE_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *ComposeServiceClient) Delete(container *ComposeService) error {
	return c.rancherClient.doResourceDelete(COMPOSE_SERVICE_TYPE, &container.Resource)
}

func (c *ComposeServiceClient) ActionActivate(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "activate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ComposeServiceClient) ActionCancelupgrade(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "cancelupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ComposeServiceClient) ActionContinueupgrade(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "continueupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ComposeServiceClient) ActionCreate(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ComposeServiceClient) ActionFinishupgrade(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "finishupgrade", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ComposeServiceClient) ActionRemove(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *ComposeServiceClient) ActionRollback(resource *ComposeService) (*Service, error) {

	resp := &Service{}

	err := c.rancherClient.doAction(COMPOSE_SERVICE_TYPE, "rollback", &resource.Resource, nil, resp)

	return resp, err
}
