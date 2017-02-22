// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cloudsigma

import (
	"net"
	"reflect"
	"testing"
)

type fakeCepgoClient struct {
	raw  []byte
	meta map[string]string
	keys map[string]interface{}
	err  error
}

func (f *fakeCepgoClient) All() (interface{}, error) {
	return f.keys, f.err
}

func (f *fakeCepgoClient) Key(key string) (interface{}, error) {
	return f.keys[key], f.err
}

func (f *fakeCepgoClient) Meta() (map[string]string, error) {
	return f.meta, f.err
}

func (f *fakeCepgoClient) FetchRaw(key string) ([]byte, error) {
	return f.raw, f.err
}

func TestServerContextWithEmptyPublicSSHKey(t *testing.T) {
	client := new(fakeCepgoClient)
	scs := NewServerContextService()
	scs.client = client
	client.raw = []byte(`{
		"meta": {
			"base64_fields": "cloudinit-user-data",
			"cloudinit-user-data": "I2Nsb3VkLWNvbmZpZwoKaG9zdG5hbWU6IGNvcmVvczE=",
			"ssh_public_key": ""
		}
	}`)
	metadata, err := scs.FetchMetadata()
	if err != nil {
		t.Error(err.Error())
	}

	if len(metadata.SSHPublicKeys) != 0 {
		t.Error("There should be no Public SSH Keys provided")
	}
}

func TestServerContextFetchMetadata(t *testing.T) {
	client := new(fakeCepgoClient)
	scs := NewServerContextService()
	scs.client = client
	client.raw = []byte(`{
		"context": true,
		"cpu": 4000,
		"cpu_model": null,
		"cpus_instead_of_cores": false,
		"enable_numa": false,
		"grantees": [],
		"hv_relaxed": false,
		"hv_tsc": false,
		"jobs": [],
		"mem": 4294967296,
		"meta": {
			"base64_fields": "cloudinit-user-data",
			"cloudinit-user-data": "I2Nsb3VkLWNvbmZpZwoKaG9zdG5hbWU6IGNvcmVvczE=",
			"ssh_public_key": "ssh-rsa AAAAB3NzaC1yc2E.../hQ5D5 john@doe"
		},
		"name": "coreos",
		"nics": [
			{
				"boot_order": null,
				"ip_v4_conf": {
					"conf": "dhcp",
					"ip": {
						"gateway": "31.171.244.1",
						"meta": {},
						"nameservers": [
							"178.22.66.167",
							"178.22.71.56",
							"8.8.8.8"
						],
						"netmask": 22,
						"tags": [],
						"uuid": "31.171.251.74"
					}
				},
				"ip_v6_conf": null,
				"mac": "22:3d:09:6b:90:f3",
				"model": "virtio",
				"vlan": null
			},
			{
				"boot_order": null,
				"ip_v4_conf": null,
				"ip_v6_conf": null,
				"mac": "22:ae:4a:fb:8f:31",
				"model": "virtio",
				"vlan": {
					"meta": {
						"description": "",
						"name": "CoreOS"
					},
					"tags": [],
					"uuid": "5dec030e-25b8-4621-a5a4-a3302c9d9619"
				}
			}
		],
		"smp": 2,
		"status": "running",
		"uuid": "20a0059b-041e-4d0c-bcc6-9b2852de48b3"
	}`)

	metadata, err := scs.FetchMetadata()
	if err != nil {
		t.Error(err.Error())
	}

	if metadata.Hostname != "coreos" {
		t.Errorf("Hostname is not 'coreos' but %s instead", metadata.Hostname)
	}

	if metadata.SSHPublicKeys["john@doe"] != "ssh-rsa AAAAB3NzaC1yc2E.../hQ5D5 john@doe" {
		t.Error("Public SSH Keys are not being read properly")
	}

	if !metadata.PublicIPv4.Equal(net.ParseIP("31.171.251.74")) {
		t.Errorf("Public IP is not 31.171.251.74 but %s instead", metadata.PublicIPv4)
	}
}

func TestServerContextFetchUserdata(t *testing.T) {
	client := new(fakeCepgoClient)
	scs := NewServerContextService()
	scs.client = client
	userdataSets := []struct {
		in  map[string]string
		err bool
		out []byte
	}{
		{map[string]string{
			"base64_fields":       "cloudinit-user-data",
			"cloudinit-user-data": "aG9zdG5hbWU6IGNvcmVvc190ZXN0",
		}, false, []byte("hostname: coreos_test")},
		{map[string]string{
			"cloudinit-user-data": "#cloud-config\\nhostname: coreos1",
		}, false, []byte("#cloud-config\\nhostname: coreos1")},
		{map[string]string{}, false, []byte{}},
	}

	for i, set := range userdataSets {
		client.meta = set.in
		got, err := scs.FetchUserdata()
		if (err != nil) != set.err {
			t.Errorf("case %d: bad error state (got %t, want %t)", i, err != nil, set.err)
		}

		if !reflect.DeepEqual(got, set.out) {
			t.Errorf("case %d: got %s, want %s", i, got, set.out)
		}
	}
}

func TestServerContextDecodingBase64UserData(t *testing.T) {
	base64Sets := []struct {
		in  string
		out bool
	}{
		{"cloudinit-user-data,foo,bar", true},
		{"bar,cloudinit-user-data,foo,bar", true},
		{"cloudinit-user-data", true},
		{"", false},
		{"foo", false},
	}

	for _, set := range base64Sets {
		userdata := map[string]string{"base64_fields": set.in}
		if isBase64Encoded("cloudinit-user-data", userdata) != set.out {
			t.Errorf("isBase64Encoded(cloudinit-user-data, %s) should be %t", userdata, set.out)
		}
	}
}
