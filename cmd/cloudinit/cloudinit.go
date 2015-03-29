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
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/configdrive"
	"github.com/coreos/coreos-cloudinit/datasource/file"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/digitalocean"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/ec2"
	"github.com/coreos/coreos-cloudinit/datasource/proc_cmdline"
	"github.com/coreos/coreos-cloudinit/datasource/url"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
	rancherNetwork "github.com/rancherio/os/cmd/network"
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
	scriptFile string
	rancherYml string
	save       bool
	execute    bool
	network    bool
	sshKeyName string
	flags      *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.StringVar(&outputDir, "dir", "/var/lib/rancher/conf", "working directory")
	flags.StringVar(&outputFile, "file", "cloud-config-processed.yml", "output cloud config file name")
	flags.StringVar(&scriptFile, "script-file", "cloud-config-script", "output cloud config script file name")
	flags.StringVar(&rancherYml, "rancher", "cloud-config-rancher.yml", "output cloud config rancher file name")
	flags.StringVar(&sshKeyName, "ssh-key-name", "rancheros-cloud-config", "SSH key name")
	flags.BoolVar(&network, "network", true, "use network based datasources")
	flags.BoolVar(&save, "save", false, "save cloud config and exit")
	flags.BoolVar(&execute, "execute", false, "execute saved cloud config")
}

func saveFiles(script *config.Script, userdataBytes []byte, cc *config.CloudConfig) error {
	var fileData []byte

	os.Remove(path.Join(outputDir, scriptFile))
	os.Remove(path.Join(outputDir, outputFile))
	os.Remove(path.Join(outputDir, rancherYml))

	if script != nil {
		fileData = userdataBytes
		output := path.Join(outputDir, scriptFile)
		log.Infof("Writing cloud-config script to %s", output)
		if err := ioutil.WriteFile(output, fileData, 500); err != nil {
			log.Errorf("Error while writing file %v", err)
			return err
		}
	}

	if data, err := yaml.Marshal(cc); err != nil {
		log.Errorf("Error while marshalling cloud config %v", err)
		return err
	} else {
		fileData = append([]byte("#cloud-config\n"), data...)
	}

	output := path.Join(outputDir, outputFile)
	log.Infof("Writing merged cloud-config to %s", output)
	if err := ioutil.WriteFile(output, fileData, 400); err != nil {
		log.Errorf("Error while writing file %v", err)
		return err
	}

	if script == nil {
		ccData := make(map[string]interface{})
		if err := yaml.Unmarshal(userdataBytes, ccData); err != nil {
			return err
		}

		if rancher, ok := ccData["rancher"]; ok {
			bytes, err := yaml.Marshal(rancher)
			if err != nil {
				return err
			}

			if err = ioutil.WriteFile(path.Join(outputDir, rancherYml), bytes, 400); err != nil {
				return err
			}
		}
	}

	return nil
}

func Main() {
	flags.Parse(rancherConfig.FilterGlobalConfig(os.Args[1:]))

	cfg, err := rancherConfig.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to read rancher config %v", err)
	}

	dss := getDatasources(cfg)
	if len(dss) == 0 {
		log.Infof("No datasources available %v", cfg.CloudInit.Datasources)
		os.Exit(0)
	}

	ds := selectDatasource(dss)
	if ds == nil {
		log.Info("No datasources found")
		os.Exit(0)
	}

	log.Infof("Fetching user-data from datasource %v", ds.Type())
	userdataBytes, err := ds.FetchUserdata()
	if err != nil {
		log.Fatalf("Failed fetching user-data from datasource: %v", err)
	}

	log.Infof("Fetching meta-data from datasource of type %v", ds.Type())
	metadata, err := ds.FetchMetadata()
	if err != nil {
		log.Infof("Failed fetching meta-data from datasource: %v", err)
		os.Exit(1)
	}

	var ccu *config.CloudConfig
	var script *config.Script
	if ud, err := initialize.ParseUserData(string(userdataBytes)); err != nil {
		log.Fatalf("Failed to parse user-data: %v\n", err)
	} else {
		switch t := ud.(type) {
		case *config.CloudConfig:
			ccu = t
		case *config.Script:
			script = t
		}
	}

	log.Info("Merging cloud-config from meta-data and user-data")
	cc := mergeConfigs(ccu, metadata)

	if save {
		err = saveFiles(script, userdataBytes, &cc)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	if len(cc.SSHAuthorizedKeys) > 0 {
		authorizeSSHKeys("rancher", cc.SSHAuthorizedKeys, sshKeyName)
	}

	for _, user := range cc.Users {
		if user.Name == "" {
			continue
		}
		if len(user.SSHAuthorizedKeys) > 0 {
			authorizeSSHKeys(user.Name, user.SSHAuthorizedKeys, sshKeyName)
		}
	}

	for _, file := range cc.WriteFiles {
		f := system.File{File: file}
		fullPath, err := system.WriteFile(&f, outputDir)
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
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
			log.Infof("Warning: user-data hostname (%s) overrides metadata hostname (%s)\n", out.Hostname, md.Hostname)
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

	if execute {
		cloudConfig := path.Join(outputDir, outputFile)
		if _, err := os.Stat(cloudConfig); os.IsNotExist(err) {
			return dss
		}

		dss = append(dss, file.NewDatasource(cloudConfig))
		return dss
	}

	for _, ds := range cfg.CloudInit.Datasources {
		parts := strings.SplitN(ds, ":", 2)

		switch parts[0] {
		case "ec2":
			if network {
				if len(parts) == 1 {
					dss = append(dss, ec2.NewDatasource(ec2.DefaultAddress))
				} else {
					dss = append(dss, ec2.NewDatasource(parts[1]))
				}
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
		case "digitalocean":
			if network {
				if len(parts) == 1 {
					dss = append(dss, digitalocean.NewDatasource(digitalocean.DefaultAddress))
				} else {
					dss = append(dss, digitalocean.NewDatasource(parts[1]))
				}
			} else {
				enableDoLinkLocal()
			}
		}
	}

	return dss
}

func enableDoLinkLocal() {
	err := rancherNetwork.ApplyNetworkConfigs(&rancherConfig.NetworkConfig{
		Interfaces: map[string]rancherConfig.InterfaceConfig{
			"eth0": {
				IPV4LL: true,
			},
		},
	})
	if err != nil {
		log.Errorf("Failed to apply link local on eth0: %v", err)
	}
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
				log.Infof("Checking availability of %q\n", s.Type())
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
