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

package proc_cmdline

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestParseCmdlineCloudConfigFound(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{
			"cloud-config-url=example.com",
			"example.com",
		},
		{
			"cloud_config_url=example.com",
			"example.com",
		},
		{
			"cloud-config-url cloud-config-url=example.com",
			"example.com",
		},
		{
			"cloud-config-url= cloud-config-url=example.com",
			"example.com",
		},
		{
			"cloud-config-url=one.example.com cloud-config-url=two.example.com",
			"two.example.com",
		},
		{
			"foo=bar cloud-config-url=example.com ping=pong",
			"example.com",
		},
	}

	for i, tt := range tests {
		output, err := findCloudConfigURL(tt.input)
		if output != tt.expect {
			t.Errorf("Test case %d failed: %s != %s", i, output, tt.expect)
		}
		if err != nil {
			t.Errorf("Test case %d produced error: %v", i, err)
		}
	}
}

func TestProcCmdlineAndFetchConfig(t *testing.T) {

	var (
		ProcCmdlineTmpl    = "foo=bar cloud-config-url=%s/config\n"
		CloudConfigContent = "#cloud-config\n"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" && r.RequestURI == "/config" {
			fmt.Fprint(w, CloudConfigContent)
		}
	}))
	defer ts.Close()

	file, err := ioutil.TempFile(os.TempDir(), "test_proc_cmdline")
	defer os.Remove(file.Name())
	if err != nil {
		t.Errorf("Test produced error: %v", err)
	}
	_, err = file.Write([]byte(fmt.Sprintf(ProcCmdlineTmpl, ts.URL)))
	if err != nil {
		t.Errorf("Test produced error: %v", err)
	}

	p := NewDatasource()
	p.Location = file.Name()
	cfg, err := p.FetchUserdata()
	if err != nil {
		t.Errorf("Test produced error: %v", err)
	}

	if string(cfg) != CloudConfigContent {
		t.Errorf("Test failed, response body: %s != %s", cfg, CloudConfigContent)
	}
}
