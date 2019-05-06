package azure

import (
	"bytes"
	"net"
	"reflect"
	"testing"

	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/datasource/metadata"
	"github.com/rancher/os/config/cloudinit/datasource/metadata/test"
)

func TestType(t *testing.T) {
	want := "azure-metadata-service"
	if kind := (MetadataService{}).Type(); kind != want {
		t.Fatalf("bad type: want %q, got %q", want, kind)
	}
}

func TestMetadataURL(t *testing.T) {
	want := "http://169.254.169.254/metadata/instance?api-version=2019-02-01&format=json"
	ms := NewDatasource("")
	if url := ms.MetadataURL(); url != want {
		t.Fatalf("bad url: want %q, got %q", want, url)
	}
}

func TestUserdataURL(t *testing.T) {
	want := "http://169.254.169.254/metadata/instance/compute/customData?api-version=2019-02-01&format=text"
	ms := NewDatasource("")
	if url := ms.UserdataURL(); url != want {
		t.Fatalf("bad url: want %q, got %q", want, url)
	}
}

func TestFetchMetadata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		metadataPath string
		resources    map[string]string
		expect       datasource.Metadata
		clientErr    error
		expectErr    error
	}{
		{
			root: "/metadata/",
			resources: map[string]string{
				"/metadata/instance?api-version=2019-02-01&format=json": `{
	"compute": {
		"azEnvironment": "AZUREPUBLICCLOUD",
		"location": "westus",
		"name": "rancheros",
		"offer": "",
		"osType": "Linux",
		"placementGroupId": "",
		"plan": {
			"name": "",
			"product": "",
			"publisher": ""
		},
		"platformFaultDomain": "0",
		"platformUpdateDomain": "0",
		"provider": "Microsoft.Compute",
		"publicKeys": [{
			"keyData":"publickey1",
			"path": "/home/rancher/.ssh/authorized_keys"
		}],
		"publisher": "",
		"resourceGroupName": "rancheros",
		"sku": "Enterprise",
		"subscriptionId": "xxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxx",
		"tags": "",
		"version": "",
		"vmId": "453945c8-3923-4366-b2d3-ea4c80e9b70e",
		"vmScaleSetName": "",
		"vmSize": "Standard_A1",
		"zone": ""
	},
	"network": {
		"interface": [{
			"ipv4": {
				"ipAddress": [{
					"privateIpAddress": "192.168.1.2",
					"publicIpAddress": "5.6.7.8"
				}],
				"subnet": [{
					"address": "192.168.1.0",
					"prefix": "24"
				}]
			},
			"ipv6": {
				"ipAddress": []
			},
			"macAddress": "002248020E1E"
		}]
	}
}
`,
			},
			expect: datasource.Metadata{
				PrivateIPv4: net.ParseIP("192.168.1.2"),
				PublicIPv4:  net.ParseIP("5.6.7.8"),
				SSHPublicKeys: map[string]string{
					"0": "publickey1",
				},
				Hostname: "rancheros",
			},
		},
	} {
		service := &MetadataService{
			Service: metadata.Service{
				Root:   tt.root,
				Client: &test.HTTPClient{Resources: tt.resources, Err: tt.clientErr},
			},
		}
		metadata, err := service.FetchMetadata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): \nwant %#v,\n got %#v", tt.resources, tt.expectErr, err)
		}
		if !reflect.DeepEqual(tt.expect, metadata) {
			t.Fatalf("bad fetch (%q): \nwant %#v,\n got %#v", tt.resources, tt.expect, metadata)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root         string
		userdataPath string
		resources    map[string]string
		userdata     []byte
		clientErr    error
		expectErr    error
	}{
		{
			root: "/metadata/",
			resources: map[string]string{
				"/metadata/instance/compute/customData?api-version=2019-02-01&format=text": "I2Nsb3VkLWNvbmZpZwpob3N0bmFtZTogcmFuY2hlcjE=",
			},
			userdata: []byte(`#cloud-config
hostname: rancher1`),
		},
	} {
		service := &MetadataService{
			Service: metadata.Service{
				Root:   tt.root,
				Client: &test.HTTPClient{Resources: tt.resources, Err: tt.clientErr},
			},
		}
		data, err := service.FetchUserdata()
		if Error(err) != Error(tt.expectErr) {
			t.Fatalf("bad error (%q): want %q, got %q", tt.resources, tt.expectErr, err)
		}
		if !bytes.Equal(data, tt.userdata) {
			t.Fatalf("bad userdata (%q): want %q, got %q", tt.resources, tt.userdata, data)
		}
	}
}

func Error(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
