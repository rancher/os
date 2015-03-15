package config

import "testing"
import "reflect"

func TestParseCmdline(t *testing.T) {
	expected := map[string]interface{}{
		"rescue":   true,
		"key1":     "value1",
		"key2":     "value2",
		"keyArray": []string{"1", "2"},
		"obj1": map[string]interface{}{
			"key3": "3value",
			"obj2": map[string]interface{}{
				"key4": true,
			},
		},
		"key5": 5,
	}

	actual := parseCmdline("a b rancher.rescue rancher.keyArray=[1,2] rancher.key1=value1 c rancher.key2=value2 rancher.obj1.key3=3value rancher.obj1.obj2.key4 rancher.key5=5")

	ok := reflect.DeepEqual(actual, expected)
	if !ok {
		t.Fatalf("%v != %v", actual, expected)
	}
}

func TestGet(t *testing.T) {
	data := map[interface{}]interface{}{
		"key": "value",
		"key2": map[interface{}]interface{}{
			"subkey": "subvalue",
			"subnum": 42,
		},
	}

	tests := map[string]interface{}{
		"key":          "value",
		"key2.subkey":  "subvalue",
		"key2.subnum":  42,
		"key2.subkey2": "",
		"foo":          "",
	}

	for k, v := range tests {
		if getOrSetVal(k, data, nil) != v {
			t.Fatalf("Expected %v, got %v, for key %s", v, getOrSetVal(k, data, nil), k)
		}
	}
}

func TestSet(t *testing.T) {
	data := map[interface{}]interface{}{
		"key": "value",
		"key2": map[interface{}]interface{}{
			"subkey": "subvalue",
			"subnum": 42,
		},
	}

	expected := map[interface{}]interface{}{
		"key": "value2",
		"key2": map[interface{}]interface{}{
			"subkey":  "subvalue2",
			"subkey2": "value",
			"subkey3": 43,
			"subnum":  42,
		},
		"key3": map[interface{}]interface{}{
			"subkey3": 44,
		},
		"key4": "value4",
	}

	tests := map[string]interface{}{
		"key":          "value2",
		"key2.subkey":  "subvalue2",
		"key2.subkey2": "value",
		"key2.subkey3": 43,
		"key3.subkey3": 44,
		"key4":         "value4",
	}

	for k, v := range tests {
		getOrSetVal(k, data, v)
		if getOrSetVal(k, data, nil) != v {
			t.Fatalf("Expected %v, got %v, for key %s", v, getOrSetVal(k, data, nil), k)
		}
	}

	if !reflect.DeepEqual(data, expected) {
		t.Fatalf("Expected %v, got %v", expected, data)
	}
}
