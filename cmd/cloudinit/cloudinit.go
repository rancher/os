// Copyright 2015 CoreOS, Inc.
// Copyright 2015 Rancher Labs, Inc.
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

package cloudinit

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/config/validate"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/configdrive"
	"github.com/coreos/coreos-cloudinit/datasource/file"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/ec2"
	"github.com/coreos/coreos-cloudinit/datasource/proc_cmdline"
	"github.com/coreos/coreos-cloudinit/datasource/url"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
	rancherConfig "github.com/rancherio/os/config"
	"gopkg.in/yaml.v2"
)

const (
	datasourceInterval    = 100 * time.Millisecond
	datasourceMaxInterval = 30 * time.Second
	datasourceTimeout     = 5 * time.Minute
)

var (
	outputDir  string
	outputFile string
	save       bool
	sshKeyName string
)

func init() {
	flag.StringVar(&outputDir, "dir", "/var/lib/rancher/conf", "working directory")
	flag.StringVar(&outputFile, "file", "/var/lib/rancher/conf/cloud-config.yml", "cloud config file name")
	flag.StringVar(&sshKeyName, "ssh-key-name", "rancheros-cloud-config", "SSH key name")
	flag.BoolVar(&save, "save", false, "save cloud config and exit")
}

func Main() {
	flag.Parse()

	cfg, err := rancherConfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to read rancher config %v", err)
	}

	dss := getDatasources(cfg)
	if len(dss) == 0 {
		os.Exit(0)
	}

	ds := selectDatasource(dss)
	if ds == nil {
		log.Info("No datasources found")
		os.Exit(0)
	}

	log.Info("Fetching user-data from datasource %s", ds.Type())
	userdataBytes, err := ds.FetchUserdata()
	if err != nil {
		log.Fatalf("Failed fetching user-data from datasource: %v", err)
	}

	if report, err := validate.Validate(userdataBytes); err == nil {
		fail := false
		for _, e := range report.Entries() {
			log.Error(e)
			fail = true
		}
		if fail {
			os.Exit(1)
		}
	} else {
		log.Fatalf("Failed while validating user_data (%v)", err)
	}

	fmt.Printf("Fetching meta-data from datasource of type %v", ds.Type())
	metadata, err := ds.FetchMetadata()
	if err != nil {
		fmt.Printf("Failed fetching meta-data from datasource: %v", err)
		os.Exit(1)
	}

	// Apply environment to user-data
	env := initialize.NewEnvironment("/", ds.ConfigRoot(), outputDir, sshKeyName, metadata)
	userdata := env.Apply(string(userdataBytes))

	var ccu *config.CloudConfig
	var script *config.Script
	if ud, err := initialize.ParseUserData(userdata); err != nil {
		log.Fatalf("Failed to parse user-data: %v\n", err)
	} else {
		switch t := ud.(type) {
		case *config.CloudConfig:
			ccu = t
		case *config.Script:
			script = t
		}
	}

	fmt.Println("Merging cloud-config from meta-data and user-data")
	cc := mergeConfigs(ccu, metadata)

	if save {
		var fileData []byte

		if script != nil {
			fileData = userdataBytes
		} else {
			if data, err := yaml.Marshal(cc); err != nil {
				log.Fatalf("Error while marshalling cloud config %v", err)
			} else {
				fileData = data
			}
		}

		if err := ioutil.WriteFile(outputFile, fileData, 400); err != nil {
			log.Fatalf("Error while writing file %v", err)
		}

		os.Exit(0)
	}

	if err = initialize.Apply(cc, []network.InterfaceGenerator{}, env); err != nil {
		log.Fatalf("Failed to apply cloud-config: %v", err)
	}

	if script != nil {
		if err = runScript(*script, env); err != nil {
			log.Fatalf("Failed to run script: %v", err)
		}
	}
}

// mergeConfigs merges certain options from md (meta-data from the datasource)
// onto cc (a CloudConfig derived from user-data), if they are not already set
// on cc (i.e. user-data always takes precedence)
func mergeConfigs(cc *config.CloudConfig, md datasource.Metadata) (out config.CloudConfig) {
	if cc != nil {
		out = *cc
	}

	if md.Hostname != "" {
		if out.Hostname != "" {
			fmt.Printf("Warning: user-data hostname (%s) overrides metadata hostname (%s)\n", out.Hostname, md.Hostname)
		} else {
			out.Hostname = md.Hostname
		}
	}
	for _, key := range md.SSHPublicKeys {
		out.SSHAuthorizedKeys = append(out.SSHAuthorizedKeys, key)
	}
	return
}

// getDatasources creates a slice of possible Datasources for cloudinit based
// on the different source command-line flags.
func getDatasources(cfg *rancherConfig.Config) []datasource.Datasource {
	dss := make([]datasource.Datasource, 0, 5)

	for _, ds := range cfg.CloudInit.Datasources {
		parts := strings.SplitN(ds, ":", 1)

		switch parts[0] {
		case "ec2":
			if len(parts) == 1 {
				dss = append(dss, ec2.NewDatasource(ec2.DefaultAddress))
			} else {
				dss = append(dss, ec2.NewDatasource(parts[1]))
			}
		case "file":
			if len(parts) == 2 {
				dss = append(dss, file.NewDatasource(parts[1]))
			}
		case "url":
			if len(parts) == 2 {
				dss = append(dss, url.NewDatasource(parts[1]))
			}
		case "cmdline":
			if len(parts) == 2 {
				dss = append(dss, proc_cmdline.NewDatasource())
			}
		case "configdrive":
			if len(parts) == 2 {
				dss = append(dss, configdrive.NewDatasource(parts[1]))
			}
		}
	}

	return dss
}

// selectDatasource attempts to choose a valid Datasource to use based on its
// current availability. The first Datasource to report to be available is
// returned. Datasources will be retried if possible if they are not
// immediately available. If all Datasources are permanently unavailable or
// datasourceTimeout is reached before one becomes available, nil is returned.
func selectDatasource(sources []datasource.Datasource) datasource.Datasource {
	ds := make(chan datasource.Datasource)
	stop := make(chan struct{})
	var wg sync.WaitGroup

	for _, s := range sources {
		wg.Add(1)
		go func(s datasource.Datasource) {
			defer wg.Done()

			duration := datasourceInterval
			for {
				fmt.Printf("Checking availability of %q\n", s.Type())
				if s.IsAvailable() {
					ds <- s
					return
				} else if !s.AvailabilityChanges() {
					return
				}
				select {
				case <-stop:
					return
				case <-time.After(duration):
					duration = pkg.ExpBackoff(duration, datasourceMaxInterval)
				}
			}
		}(s)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	var s datasource.Datasource
	select {
	case s = <-ds:
	case <-done:
	case <-time.After(datasourceTimeout):
	}

	close(stop)
	return s
}

// TODO(jonboulle): this should probably be refactored and moved into a different module
func runScript(script config.Script, env *initialize.Environment) error {
	err := initialize.PrepWorkspace(env.Workspace())
	if err != nil {
		fmt.Printf("Failed preparing workspace: %v\n", err)
		return err
	}
	path, err := initialize.PersistScriptInWorkspace(script, env.Workspace())
	if err == nil {
		var name string
		name, err = system.ExecuteScript(path)
		initialize.PersistUnitNameInWorkspace(name, env.Workspace())
	}
	return err
}
