package cepgo

import (
	"encoding/json"
	"testing"
)

func fetchMock(key string) ([]byte, error) {
	context := []byte(`{
		"context": true,
		"cpu": 4000,
		"cpu_model": null,
		"cpus_instead_of_cores": false,
		"enable_numa": false,
		"global_context": {
			"some_global_key": "some_global_val"
		},
		"grantees": [],
		"hv_relaxed": false,
		"hv_tsc": false,
		"jobs": [],
		"mem": 4294967296,
		"meta": {
			"base64_fields": "cloudinit-user-data",
			"cloudinit-user-data": "I2Nsb3VkLWNvbmZpZwoKaG9zdG5hbWU6IGNvcmVvczE=",
			"ssh_public_key": "ssh-rsa AAAAB2NzaC1yc2E.../hQ5D5 john@doe"
		},
		"name": "coreos",
		"nics": [
			{
				"runtime": {
					"interface_type": "public",
					"ip_v4": {
						"uuid": "31.171.251.74"
					},
					"ip_v6": null
				},
				"vlan": null
			}
		],
		"smp": 2,
		"status": "running",
		"uuid": "20a0059b-041e-4d0c-bcc6-9b2852de48b3"
	}`)

	if key == "" {
		return context, nil
	}

	var marshalledContext map[string]interface{}

	err := json.Unmarshal(context, &marshalledContext)
	if err != nil {
		return nil, err
	}

	if key[0] == '/' {
		key = key[1:]
	}
	if key[len(key)-1] == '/' {
		key = key[:len(key)-1]
	}

	return json.Marshal(marshalledContext[key])
}

func TestAll(t *testing.T) {
	cepgo := NewCepgoFetcher(fetchMock)

	result, err := cepgo.All()
	if err != nil {
		t.Error(err)
	}

	for _, key := range []string{"meta", "name", "uuid", "global_context"} {
		if _, ok := result.(map[string]interface{})[key]; !ok {
			t.Errorf("%s not in all keys", key)
		}
	}
}

func TestKey(t *testing.T) {
	cepgo := NewCepgoFetcher(fetchMock)

	result, err := cepgo.Key("uuid")
	if err != nil {
		t.Error(err)
	}

	if _, ok := result.(string); !ok {
		t.Errorf("%#v\n", result)

		t.Error("Fetching the uuid did not return a string")
	}
}

func TestMeta(t *testing.T) {
	cepgo := NewCepgoFetcher(fetchMock)

	meta, err := cepgo.Meta()
	if err != nil {
		t.Errorf("%#v\n", meta)
		t.Error(err)
	}

	if _, ok := meta["ssh_public_key"]; !ok {
		t.Error("ssh_public_key is not in the meta")
	}
}

func TestGlobalContext(t *testing.T) {
	cepgo := NewCepgoFetcher(fetchMock)

	result, err := cepgo.GlobalContext()
	if err != nil {
		t.Error(err)
	}

	if _, ok := result["some_global_key"]; !ok {
		t.Error("some_global_key is not in the global context")
	}
}
