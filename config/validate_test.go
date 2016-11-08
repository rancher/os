package config

import (
	"fmt"
	"strings"
	"testing"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/rancher/os/util"
)

func testValidate(t *testing.T, cfg []byte, contains string) {
	validationErrors, err := Validate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if contains == "" && len(validationErrors.Errors()) != 0 {
		t.Fail()
	}
	if !strings.Contains(fmt.Sprint(validationErrors.Errors()), contains) {
		t.Fail()
	}
}

func TestValidate(t *testing.T) {
	testValidate(t, []byte("{}"), "")
	testValidate(t, []byte(`rancher:
  log: true
`), "")
	testValidate(t, []byte("bad_key: {}"), "Additional property bad_key is not allowed")
	testValidate(t, []byte("rancher: []"), "rancher: Invalid type. Expected: object, given: array")

	var fullConfig map[string]interface{}
	if err := util.ConvertIgnoreOmitEmpty(CloudConfig{}, &fullConfig); err != nil {
		t.Fatal(err)
	}
	fullConfigBytes, err := yaml.Marshal(fullConfig)
	if err != nil {
		t.Fatal(err)
	}
	testValidate(t, fullConfigBytes, "")
}
