package config

import (
	"testing"

	"code.google.com/p/rog-go/deepdiff"
)

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

	ok, err := deepdiff.DeepDiff(actual, expected)
	if !ok || err != nil {
		t.Fatal(err)
	}
}
