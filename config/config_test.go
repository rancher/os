package config

import (
	"testing"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	"github.com/rancher/os/util"
	"github.com/stretchr/testify/require"
)

func TestFilterKey(t *testing.T) {
	assert := require.New(t)
	data := map[interface{}]interface{}{
		"ssh_authorized_keys": []string{"pubk1", "pubk2"},
		"hostname":            "ros-test",
		"rancher": map[interface{}]interface{}{
			"ssh": map[interface{}]interface{}{
				"keys": map[interface{}]interface{}{
					"dsa":     "dsa-test1",
					"dsa-pub": "dsa-test2",
				},
			},
			"docker": map[interface{}]interface{}{
				"ca_key":  "ca_key-test3",
				"ca_cert": "ca_cert-test4",
				"args":    []string{"args_test5"},
			},
		},
	}
	expectedFiltered := map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"ssh": map[interface{}]interface{}{
				"keys": map[interface{}]interface{}{
					"dsa":     "dsa-test1",
					"dsa-pub": "dsa-test2",
				},
			},
		},
	}
	expectedRest := map[interface{}]interface{}{
		"ssh_authorized_keys": []string{"pubk1", "pubk2"},
		"hostname":            "ros-test",
		"rancher": map[interface{}]interface{}{
			"docker": map[interface{}]interface{}{
				"ca_key":  "ca_key-test3",
				"ca_cert": "ca_cert-test4",
				"args":    []string{"args_test5"},
			},
		},
	}
	filtered, rest := filterKey(data, []string{"rancher", "ssh"})
	assert.Equal(expectedFiltered, filtered)
	assert.Equal(expectedRest, rest)
}

func TestUnmarshalOrReturnString(t *testing.T) {
	assert := require.New(t)

	assert.Equal("ab", unmarshalOrReturnString("ab"))
	assert.Equal("a\nb", unmarshalOrReturnString("a\nb"))
	assert.Equal("a\n", unmarshalOrReturnString("a\n"))
	assert.Equal("\nb", unmarshalOrReturnString("\nb"))
	assert.Equal("a,b", unmarshalOrReturnString("a,b"))
	assert.Equal("a,", unmarshalOrReturnString("a,"))
	assert.Equal(",b", unmarshalOrReturnString(",b"))

	assert.Equal(int64(10), unmarshalOrReturnString("10"))
	assert.Equal(true, unmarshalOrReturnString("true"))
	assert.Equal(false, unmarshalOrReturnString("false"))

	assert.Equal([]interface{}{"a"}, unmarshalOrReturnString("[a]"))
	assert.Equal([]interface{}{"a"}, unmarshalOrReturnString("[\"a\"]"))

	assert.Equal([]interface{}{"a,"}, unmarshalOrReturnString("[\"a,\"]"))
	assert.Equal([]interface{}{" a, "}, unmarshalOrReturnString("[\" a, \"]"))
	assert.Equal([]interface{}{",a"}, unmarshalOrReturnString("[\",a\"]"))
	assert.Equal([]interface{}{" ,a "}, unmarshalOrReturnString("[\" ,a \"]"))

	assert.Equal([]interface{}{"a\n"}, unmarshalOrReturnString("[\"a\n\"]"))
	assert.Equal([]interface{}{" a\n "}, unmarshalOrReturnString("[\" a\n \"]"))
	assert.Equal([]interface{}{"\na"}, unmarshalOrReturnString("[\"\na\"]"))
	assert.Equal([]interface{}{" \na "}, unmarshalOrReturnString("[\" \na \"]"))

	assert.Equal([]interface{}{"a", "b"}, unmarshalOrReturnString("[a,b]"))
	assert.Equal([]interface{}{"a", "b"}, unmarshalOrReturnString("[\"a\",\"b\"]"))

	assert.Equal([]interface{}{"a,", "b"}, unmarshalOrReturnString("[\"a,\",b]"))
	assert.Equal([]interface{}{"a", ",b"}, unmarshalOrReturnString("[a,\",b\"]"))
	assert.Equal([]interface{}{" a, ", " ,b "}, unmarshalOrReturnString("[\" a, \",\" ,b \"]"))

	assert.Equal([]interface{}{"a\n", "b"}, unmarshalOrReturnString("[\"a\n\",b]"))
	assert.Equal([]interface{}{"a", "\nb"}, unmarshalOrReturnString("[a,\"\nb\"]"))
	assert.Equal([]interface{}{" a\n ", " \nb "}, unmarshalOrReturnString("[\" a\n \",\" \nb \"]"))

	assert.Equal([]interface{}{"a", int64(10)}, unmarshalOrReturnString("[a,10]"))
	assert.Equal([]interface{}{int64(10), "a"}, unmarshalOrReturnString("[10,a]"))

	assert.Equal([]interface{}{"a", true}, unmarshalOrReturnString("[a,true]"))
	assert.Equal([]interface{}{false, "a"}, unmarshalOrReturnString("[false,a]"))
}

func TestParseCmdline(t *testing.T) {
	assert := require.New(t)

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
	}, parseCmdline("a b rancher.key1=value1 c rancher.key2=value2"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"key": "a,b",
		},
	}, parseCmdline("rancher.key=a,b"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"key": "a\nb",
		},
	}, parseCmdline("rancher.key=a\nb"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"key": "a:b",
		},
	}, parseCmdline("rancher.key=a:b"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"key": int64(5),
		},
	}, parseCmdline("rancher.key=5"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"rescue": true,
		},
	}, parseCmdline("rancher.rescue"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"keyArray": []interface{}{int64(1), int64(2)},
		},
	}, parseCmdline("rancher.keyArray=[1,2]"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"strArray": []interface{}{"url:http://192.168.1.100/cloud-config"},
		},
	}, parseCmdline("rancher.strArray=[\"url:http://192.168.1.100/cloud-config\"]"))

	assert.Equal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"strArray": []interface{}{"url:http://192.168.1.100/cloud-config"},
		},
	}, parseCmdline("rancher.strArray=[url:http://192.168.1.100/cloud-config]"))
}

func TestGet(t *testing.T) {
	assert := require.New(t)

	data := map[interface{}]interface{}{
		"key": "value",
		"rancher": map[interface{}]interface{}{
			"key2": map[interface{}]interface{}{
				"subkey": "subvalue",
				"subnum": 42,
			},
		},
	}

	tests := map[string]interface{}{
		"key": "value",
		"rancher.key2.subkey":  "subvalue",
		"rancher.key2.subnum":  42,
		"rancher.key2.subkey2": "",
		"foo": "",
	}

	for k, v := range tests {
		val, _ := getOrSetVal(k, data, nil)
		assert.Equal(v, val)
	}
}

func TestSet(t *testing.T) {
	assert := require.New(t)

	data := map[interface{}]interface{}{
		"key": "value",
		"rancher": map[interface{}]interface{}{
			"key2": map[interface{}]interface{}{
				"subkey": "subvalue",
				"subnum": 42,
			},
		},
	}

	expected := map[interface{}]interface{}{
		"key": "value2",
		"rancher": map[interface{}]interface{}{
			"key2": map[interface{}]interface{}{
				"subkey":  "subvalue2",
				"subkey2": "value",
				"subkey3": 43,
				"subnum":  42,
			},
			"key3": map[interface{}]interface{}{
				"subkey3": 44,
			},
		},
		"key4": "value4",
	}

	tests := map[string]interface{}{
		"key": "value2",
		"rancher.key2.subkey":  "subvalue2",
		"rancher.key2.subkey2": "value",
		"rancher.key2.subkey3": 43,
		"rancher.key3.subkey3": 44,
		"key4":                 "value4",
	}

	for k, v := range tests {
		_, tData := getOrSetVal(k, data, v)
		val, _ := getOrSetVal(k, tData, nil)
		data = tData
		assert.Equal(v, val)
	}

	assert.Equal(expected, data)
}

type OuterData struct {
	One Data `yaml:"one"`
}

type Data struct {
	Two   bool `yaml:"two"`
	Three bool `yaml:"three"`
}

func TestMapMerge(t *testing.T) {
	assert := require.New(t)
	one := `
one:
  two: true`
	two := `
one:
  three: true`

	data := map[string]map[string]bool{}
	yaml.Unmarshal([]byte(one), &data)
	yaml.Unmarshal([]byte(two), &data)

	assert.NotNil(data["one"])
	assert.True(data["one"]["three"])
	assert.False(data["one"]["two"])

	data2 := &OuterData{}
	yaml.Unmarshal([]byte(one), data2)
	yaml.Unmarshal([]byte(two), data2)

	assert.True(data2.One.Three)
	assert.True(data2.One.Two)
}

func TestUserDocker(t *testing.T) {
	assert := require.New(t)

	config := &CloudConfig{
		Rancher: RancherConfig{
			Docker: DockerConfig{
				TLS: true,
			},
		},
	}

	bytes, err := yaml.Marshal(config)
	assert.Nil(err)

	config = &CloudConfig{}
	assert.False(config.Rancher.Docker.TLS)
	err = yaml.Unmarshal(bytes, config)
	assert.Nil(err)
	assert.True(config.Rancher.Docker.TLS)

	data := map[interface{}]interface{}{}
	err = util.Convert(config, &data)
	assert.Nil(err)

	val, ok := data["rancher"].(map[interface{}]interface{})["docker"]
	assert.True(ok)

	m, ok := val.(map[interface{}]interface{})
	assert.True(ok)
	v, ok := m["tls"]
	assert.True(ok)
	assert.True(v.(bool))

}
