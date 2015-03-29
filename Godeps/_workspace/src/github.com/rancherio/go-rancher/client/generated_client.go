package client

type RancherClient struct {
    RancherBaseClient
	
    Subscribe SubscribeOperations
    Publish PublishOperations
    RestartPolicy RestartPolicyOperations
    LoadBalancerHealthCheck LoadBalancerHealthCheckOperations
    LoadBalancerCookieStickinessPolicy LoadBalancerCookieStickinessPolicyOperations
    LoadBalancerAppCookieStickinessPolicy LoadBalancerAppCookieStickinessPolicyOperations
    GlobalLoadBalancerPolicy GlobalLoadBalancerPolicyOperations
    GlobalLoadBalancerHealthCheck GlobalLoadBalancerHealthCheckOperations
    Container ContainerOperations
    ApiKey ApiKeyOperations
    InstanceStop InstanceStopOperations
    InstanceConsole InstanceConsoleOperations
    InstanceConsoleInput InstanceConsoleInputOperations
    IpAddressAssociateInput IpAddressAssociateInputOperations
    Project ProjectOperations
    AddRemoveLoadBalancerListenerInput AddRemoveLoadBalancerListenerInputOperations
    AddRemoveLoadBalancerTargetInput AddRemoveLoadBalancerTargetInputOperations
    AddLoadBalancerInput AddLoadBalancerInputOperations
    RemoveLoadBalancerInput RemoveLoadBalancerInputOperations
    AddRemoveLoadBalancerHostInput AddRemoveLoadBalancerHostInputOperations
    SetLoadBalancerListenersInput SetLoadBalancerListenersInputOperations
    SetLoadBalancerTargetsInput SetLoadBalancerTargetsInputOperations
    SetLoadBalancerHostsInput SetLoadBalancerHostsInputOperations
    Cluster ClusterOperations
    AddRemoveClusterHostInput AddRemoveClusterHostInputOperations
    RegistryCredential RegistryCredentialOperations
    Registry RegistryOperations
    Account AccountOperations
    Agent AgentOperations
    Certificate CertificateOperations
    ConfigItem ConfigItemOperations
    ConfigItemStatus ConfigItemStatusOperations
    ContainerEvent ContainerEventOperations
    Credential CredentialOperations
    Databasechangelog DatabasechangelogOperations
    Databasechangeloglock DatabasechangeloglockOperations
    ExternalHandler ExternalHandlerOperations
    ExternalHandlerExternalHandlerProcessMap ExternalHandlerExternalHandlerProcessMapOperations
    ExternalHandlerProcess ExternalHandlerProcessOperations
    GlobalLoadBalancer GlobalLoadBalancerOperations
    Host HostOperations
    Image ImageOperations
    Instance InstanceOperations
    InstanceLink InstanceLinkOperations
    IpAddress IpAddressOperations
    LoadBalancer LoadBalancerOperations
    LoadBalancerConfig LoadBalancerConfigOperations
    LoadBalancerListener LoadBalancerListenerOperations
    LoadBalancerTarget LoadBalancerTargetOperations
    Mount MountOperations
    Network NetworkOperations
    PhysicalHost PhysicalHostOperations
    Port PortOperations
    ProcessExecution ProcessExecutionOperations
    ProcessInstance ProcessInstanceOperations
    Setting SettingOperations
    StoragePool StoragePoolOperations
    Task TaskOperations
    TaskInstance TaskInstanceOperations
    Volume VolumeOperations
    TypeDocumentation TypeDocumentationOperations
    ContainerExec ContainerExecOperations
    ContainerLogs ContainerLogsOperations
    HostAccess HostAccessOperations
    ActiveSetting ActiveSettingOperations
    ExtensionImplementation ExtensionImplementationOperations
    ExtensionPoint ExtensionPointOperations
    ProcessDefinition ProcessDefinitionOperations
    ResourceDefinition ResourceDefinitionOperations
    Githubconfig GithubconfigOperations
    StatsAccess StatsAccessOperations
    VirtualboxConfig VirtualboxConfigOperations
    DigitaloceanConfig DigitaloceanConfigOperations
    Machine MachineOperations
    Register RegisterOperations
    RegistrationToken RegistrationTokenOperations
}

func constructClient() *RancherClient {
	client := &RancherClient{
		RancherBaseClient: RancherBaseClient{
			Types: map[string]Schema{},
		},
	}

    
    client.Subscribe = newSubscribeClient(client)
    client.Publish = newPublishClient(client)
    client.RestartPolicy = newRestartPolicyClient(client)
    client.LoadBalancerHealthCheck = newLoadBalancerHealthCheckClient(client)
    client.LoadBalancerCookieStickinessPolicy = newLoadBalancerCookieStickinessPolicyClient(client)
    client.LoadBalancerAppCookieStickinessPolicy = newLoadBalancerAppCookieStickinessPolicyClient(client)
    client.GlobalLoadBalancerPolicy = newGlobalLoadBalancerPolicyClient(client)
    client.GlobalLoadBalancerHealthCheck = newGlobalLoadBalancerHealthCheckClient(client)
    client.Container = newContainerClient(client)
    client.ApiKey = newApiKeyClient(client)
    client.InstanceStop = newInstanceStopClient(client)
    client.InstanceConsole = newInstanceConsoleClient(client)
    client.InstanceConsoleInput = newInstanceConsoleInputClient(client)
    client.IpAddressAssociateInput = newIpAddressAssociateInputClient(client)
    client.Project = newProjectClient(client)
    client.AddRemoveLoadBalancerListenerInput = newAddRemoveLoadBalancerListenerInputClient(client)
    client.AddRemoveLoadBalancerTargetInput = newAddRemoveLoadBalancerTargetInputClient(client)
    client.AddLoadBalancerInput = newAddLoadBalancerInputClient(client)
    client.RemoveLoadBalancerInput = newRemoveLoadBalancerInputClient(client)
    client.AddRemoveLoadBalancerHostInput = newAddRemoveLoadBalancerHostInputClient(client)
    client.SetLoadBalancerListenersInput = newSetLoadBalancerListenersInputClient(client)
    client.SetLoadBalancerTargetsInput = newSetLoadBalancerTargetsInputClient(client)
    client.SetLoadBalancerHostsInput = newSetLoadBalancerHostsInputClient(client)
    client.Cluster = newClusterClient(client)
    client.AddRemoveClusterHostInput = newAddRemoveClusterHostInputClient(client)
    client.RegistryCredential = newRegistryCredentialClient(client)
    client.Registry = newRegistryClient(client)
    client.Account = newAccountClient(client)
    client.Agent = newAgentClient(client)
    client.Certificate = newCertificateClient(client)
    client.ConfigItem = newConfigItemClient(client)
    client.ConfigItemStatus = newConfigItemStatusClient(client)
    client.ContainerEvent = newContainerEventClient(client)
    client.Credential = newCredentialClient(client)
    client.Databasechangelog = newDatabasechangelogClient(client)
    client.Databasechangeloglock = newDatabasechangeloglockClient(client)
    client.ExternalHandler = newExternalHandlerClient(client)
    client.ExternalHandlerExternalHandlerProcessMap = newExternalHandlerExternalHandlerProcessMapClient(client)
    client.ExternalHandlerProcess = newExternalHandlerProcessClient(client)
    client.GlobalLoadBalancer = newGlobalLoadBalancerClient(client)
    client.Host = newHostClient(client)
    client.Image = newImageClient(client)
    client.Instance = newInstanceClient(client)
    client.InstanceLink = newInstanceLinkClient(client)
    client.IpAddress = newIpAddressClient(client)
    client.LoadBalancer = newLoadBalancerClient(client)
    client.LoadBalancerConfig = newLoadBalancerConfigClient(client)
    client.LoadBalancerListener = newLoadBalancerListenerClient(client)
    client.LoadBalancerTarget = newLoadBalancerTargetClient(client)
    client.Mount = newMountClient(client)
    client.Network = newNetworkClient(client)
    client.PhysicalHost = newPhysicalHostClient(client)
    client.Port = newPortClient(client)
    client.ProcessExecution = newProcessExecutionClient(client)
    client.ProcessInstance = newProcessInstanceClient(client)
    client.Setting = newSettingClient(client)
    client.StoragePool = newStoragePoolClient(client)
    client.Task = newTaskClient(client)
    client.TaskInstance = newTaskInstanceClient(client)
    client.Volume = newVolumeClient(client)
    client.TypeDocumentation = newTypeDocumentationClient(client)
    client.ContainerExec = newContainerExecClient(client)
    client.ContainerLogs = newContainerLogsClient(client)
    client.HostAccess = newHostAccessClient(client)
    client.ActiveSetting = newActiveSettingClient(client)
    client.ExtensionImplementation = newExtensionImplementationClient(client)
    client.ExtensionPoint = newExtensionPointClient(client)
    client.ProcessDefinition = newProcessDefinitionClient(client)
    client.ResourceDefinition = newResourceDefinitionClient(client)
    client.Githubconfig = newGithubconfigClient(client)
    client.StatsAccess = newStatsAccessClient(client)
    client.VirtualboxConfig = newVirtualboxConfigClient(client)
    client.DigitaloceanConfig = newDigitaloceanConfigClient(client)
    client.Machine = newMachineClient(client)
    client.Register = newRegisterClient(client)
    client.RegistrationToken = newRegistrationTokenClient(client) 


	return client
}

func NewRancherClient(opts *ClientOpts) (*RancherClient, error) {
    client := constructClient()
        
    err := setupRancherBaseClient(&client.RancherBaseClient, opts)
    if err != nil {
        return nil, err
    }

    return client, nil
}
