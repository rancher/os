package azure

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/burmilla/os/config/cloudinit/config"
	"github.com/burmilla/os/config/cloudinit/datasource"
	"github.com/burmilla/os/config/cloudinit/datasource/metadata"
)

const (
	metadataHeader   = "true"
	metadataVersion  = "2019-02-01"
	metadataEndpoint = "http://169.254.169.254/metadata/"
)

type MetadataService struct {
	metadata.Service
}

func NewDatasource(root string) *MetadataService {
	if root == "" {
		root = metadataEndpoint
	}
	return &MetadataService{metadata.NewDatasource(root, "instance?api-version="+metadataVersion+"&format=json", "", "", assembleHeader())}
}

func (ms MetadataService) ConfigRoot() string {
	return ms.Root + "instance"
}

func (ms MetadataService) AvailabilityChanges() bool {
	// TODO: if it can't find the network, maybe we can start it?
	return false
}

func (ms MetadataService) FetchMetadata() (datasource.Metadata, error) {
	d, err := ms.FetchData(ms.MetadataURL())
	if err != nil {
		return datasource.Metadata{}, err
	}
	type Plan struct {
		Name      string `json:"name,omitempty"`
		Product   string `json:"product,omitempty"`
		Publisher string `json:"publisher,omitempty"`
	}
	type PublicKey struct {
		KeyData string `json:"keyData,omitempty"`
		Path    string `json:"path,omitempty"`
	}
	type Compute struct {
		AZEnvironment        string      `json:"azEnvironment,omitempty"`
		CustomData           string      `json:"customData,omitempty"`
		Location             string      `json:"location,omitempty"`
		Name                 string      `json:"name,omitempty"`
		Offer                string      `json:"offer,omitempty"`
		OSType               string      `json:"osType,omitempty"`
		PlacementGroupID     string      `json:"placementGroupId,omitempty"`
		Plan                 Plan        `json:"plan,omitempty"`
		PlatformFaultDomain  string      `json:"platformFaultDomain,omitempty"`
		PlatformUpdateDomain string      `json:"platformUpdateDomain,omitempty"`
		Provider             string      `json:"provider,omitempty"`
		PublicKeys           []PublicKey `json:"publicKeys,omitempty"`
		Publisher            string      `json:"publisher,omitempty"`
		ResourceGroupName    string      `json:"resourceGroupName,omitempty"`
		SKU                  string      `json:"sku,omitempty"`
		SubscriptionID       string      `json:"subscriptionId,omitempty"`
		Tags                 string      `json:"tags,omitempty"`
		Version              string      `json:"version,omitempty"`
		VMID                 string      `json:"vmId,omitempty"`
		VMScaleSetName       string      `json:"vmScaleSetName,omitempty"`
		VMSize               string      `json:"vmSize,omitempty"`
		Zone                 string      `json:"zone,omitempty"`
	}
	type IPAddress struct {
		PrivateIPAddress string `json:"privateIpAddress,omitempty"`
		PublicIPAddress  string `json:"publicIpAddress,omitempty"`
	}
	type Subnet struct {
		Address string `json:"address,omitempty"`
		Prefix  string `json:"prefix,omitempty"`
	}
	type IPV4 struct {
		IPAddress []IPAddress `json:"ipAddress,omitempty"`
		Subnet    []Subnet    `json:"subnet,omitempty"`
	}
	type IPV6 struct {
		IPAddress []IPAddress `json:"ipAddress,omitempty"`
	}
	type Interface struct {
		IPV4       IPV4   `json:"ipv4,omitempty"`
		IPV6       IPV6   `json:"ipv6,omitempty"`
		MacAddress string `json:"macAddress,omitempty"`
	}
	type Network struct {
		Interface []Interface `json:"interface,omitempty"`
	}
	type Instance struct {
		Compute Compute `json:"compute,omitempty"`
		Network Network `json:"network,omitempty"`
	}
	instance := &Instance{}
	if err := json.Unmarshal(d, instance); err != nil {
		return datasource.Metadata{}, err
	}
	m := datasource.Metadata{
		Hostname:      instance.Compute.Name,
		SSHPublicKeys: make(map[string]string, 0),
	}
	if len(instance.Network.Interface) > 0 {
		if len(instance.Network.Interface[0].IPV4.IPAddress) > 0 {
			m.PublicIPv4 = net.ParseIP(instance.Network.Interface[0].IPV4.IPAddress[0].PublicIPAddress)
			m.PrivateIPv4 = net.ParseIP(instance.Network.Interface[0].IPV4.IPAddress[0].PrivateIPAddress)
		}
		if len(instance.Network.Interface[0].IPV6.IPAddress) > 0 {
			m.PublicIPv6 = net.ParseIP(instance.Network.Interface[0].IPV6.IPAddress[0].PublicIPAddress)
			m.PrivateIPv6 = net.ParseIP(instance.Network.Interface[0].IPV6.IPAddress[0].PrivateIPAddress)
		}
	}
	for i, k := range instance.Compute.PublicKeys {
		m.SSHPublicKeys[strconv.Itoa(i)] = k.KeyData
	}
	return m, nil
}

func (ms MetadataService) FetchUserdata() ([]byte, error) {
	d, err := ms.FetchData(ms.UserdataURL())
	if err != nil {
		return []byte{}, err
	}
	return config.DecodeBase64Content(string(d))
}

func (ms MetadataService) Type() string {
	return "azure-metadata-service"
}

func (ms MetadataService) MetadataURL() string {
	// metadata: http://169.254.169.254/metadata/instance?api-version=2019-02-01&format=json
	return ms.Root + "instance?api-version=" + metadataVersion + "&format=json"
}

func (ms MetadataService) UserdataURL() string {
	// userdata: http://169.254.169.254/metadata/instance/compute/customData?api-version=2019-02-01&format=text
	return ms.Root + "instance/compute/customData?api-version=" + metadataVersion + "&format=text"
}

func assembleHeader() http.Header {
	h := http.Header{}
	h.Add("Metadata", metadataHeader)
	return h
}
