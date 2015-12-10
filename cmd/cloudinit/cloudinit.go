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
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/configdrive"
	"github.com/coreos/coreos-cloudinit/datasource/file"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/digitalocean"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/ec2"
	"github.com/coreos/coreos-cloudinit/datasource/proc_cmdline"
	"github.com/coreos/coreos-cloudinit/datasource/url"
	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
	"github.com/rancher/netconf"
	rancherConfig "github.com/rancher/os/config"
)

const (
	datasourceInterval    = 100 * time.Millisecond
	datasourceMaxInterval = 30 * time.Second
	datasourceTimeout     = 5 * time.Minute
	sshKeyName            = "rancheros-cloud-config"
)

var (
	save    bool
	execute bool
	network bool
	flags   *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.BoolVar(&network, "network", true, "use network based datasources")
	flags.BoolVar(&save, "save", false, "save cloud config and exit")
	flags.BoolVar(&execute, "execute", false, "execute saved cloud config")
}

func saveFiles(cloudConfigBytes, scriptBytes []byte, metadata datasource.Metadata) error {
	os.MkdirAll(rancherConfig.CloudConfigDir, os.ModeDir|0600)
	os.Remove(rancherConfig.CloudConfigScriptFile)
	os.Remove(rancherConfig.CloudConfigBootFile)
	os.Remove(rancherConfig.MetaDataFile)

	if len(scriptBytes) > 0 {
		log.Infof("Writing to %s", rancherConfig.CloudConfigScriptFile)
		if err := ioutil.WriteFile(rancherConfig.CloudConfigScriptFile, scriptBytes, 500); err != nil {
			log.Errorf("Error while writing file %s: %v", rancherConfig.CloudConfigScriptFile, err)
			return err
		}
	}

	if err := ioutil.WriteFile(rancherConfig.CloudConfigBootFile, cloudConfigBytes, 400); err != nil {
		return err
	}
	log.Infof("Written to %s:\n%s", rancherConfig.CloudConfigBootFile, string(cloudConfigBytes))

	metaDataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(rancherConfig.MetaDataFile, metaDataBytes, 400); err != nil {
		return err
	}
	log.Infof("Written to %s:\n%s", rancherConfig.MetaDataFile, string(metaDataBytes))

	return nil
}

func currentDatasource() (datasource.Datasource, error) {
	cfg, err := rancherConfig.LoadConfig()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to read rancher config")
		return nil, err
	}

	dss := getDatasources(cfg)
	if len(dss) == 0 {
		return nil, nil
	}

	ds := selectDatasource(dss)
	return ds, nil
}

func saveCloudConfig() error {
	userDataBytes, metadata, err := fetchUserData()
	if err != nil {
		return err
	}

	userData := string(userDataBytes)
	scriptBytes := []byte{}

	if config.IsScript(userData) {
		scriptBytes = userDataBytes
		userDataBytes = []byte{}
	} else if isCompose(userData) {
		if userDataBytes, err = composeToCloudConfig(userDataBytes); err != nil {
			log.Errorf("Failed to convert compose to cloud-config syntax: %v", err)
			return err
		}
	} else if config.IsCloudConfig(userData) {
		if _, err := rancherConfig.ReadConfig(userDataBytes, false); err != nil {
			log.WithFields(log.Fields{"cloud-config": userData, "err": err}).Warn("Failed to parse cloud-config, not saving.")
			userDataBytes = []byte{}
		}
	} else {
		log.Errorf("Unrecognized user-data\n%s", userData)
		userDataBytes = []byte{}
	}

	if _, err := rancherConfig.ReadConfig(userDataBytes, false); err != nil {
		log.WithFields(log.Fields{"cloud-config": userData, "err": err}).Warn("Failed to parse cloud-config")
		return errors.New("Failed to parse cloud-config")
	}

	return saveFiles(userDataBytes, scriptBytes, metadata)
}

func fetchUserData() ([]byte, datasource.Metadata, error) {
	var metadata datasource.Metadata
	ds, err := currentDatasource()
	if err != nil || ds == nil {
		log.Errorf("Failed to select datasource: %v", err)
		return nil, metadata, err
	}
	log.Infof("Fetching user-data from datasource %v", ds.Type())
	userDataBytes, err := ds.FetchUserdata()
	if err != nil {
		log.Errorf("Failed fetching user-data from datasource: %v", err)
		return nil, metadata, err
	}
	log.Infof("Fetching meta-data from datasource of type %v", ds.Type())
	metadata, err = ds.FetchMetadata()
	if err != nil {
		log.Errorf("Failed fetching meta-data from datasource: %v", err)
		return nil, metadata, err
	}
	return userDataBytes, metadata, nil
}

func SetHostname(cc *rancherConfig.CloudConfig) (string, error) {
	name, _ := os.Hostname()
	if cc.Hostname != "" {
		name = cc.Hostname
	}
	if name != "" {
		//set hostname
		if err := syscall.Sethostname([]byte(name)); err != nil {
			log.WithFields(log.Fields{"err": err, "hostname": name}).Error("Error setting hostname")
			return "", err
		}
	}

	return name, nil
}

func executeCloudConfig() error {
	cc, err := rancherConfig.LoadConfig()
	if err != nil {
		return err
	}

	if _, err := SetHostname(cc); err != nil {
		return err
	}

	if len(cc.SSHAuthorizedKeys) > 0 {
		authorizeSSHKeys("rancher", cc.SSHAuthorizedKeys, sshKeyName)
		authorizeSSHKeys("docker", cc.SSHAuthorizedKeys, sshKeyName)
	}

	for _, file := range cc.WriteFiles {
		f := system.File{File: file}
		fullPath, err := system.WriteFile(&f, "/")
		if err != nil {
			log.WithFields(log.Fields{"err": err, "path": fullPath}).Error("Error writing file")
			continue
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
	}

	return nil
}

func Main() {
	flags.Parse(os.Args[1:])

	log.Infof("Running cloud-init: save=%v, execute=%v", save, execute)

	if save {
		err := saveCloudConfig()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed to save cloud-config")
		}
	}

	if execute {
		err := executeCloudConfig()
		if err != nil {
			log.WithFields(log.Fields{"err": err}).Error("Failed to execute cloud-config")
		}
	}
}

// getDatasources creates a slice of possible Datasources for cloudinit based
// on the different source command-line flags.
func getDatasources(cfg *rancherConfig.CloudConfig) []datasource.Datasource {
	dss := make([]datasource.Datasource, 0, 5)

	for _, ds := range cfg.Rancher.CloudInit.Datasources {
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
			if network {
				if len(parts) == 2 {
					dss = append(dss, url.NewDatasource(parts[1]))
				}
			}
		case "cmdline":
			if network {
				if len(parts) == 1 {
					dss = append(dss, proc_cmdline.NewDatasource())
				}
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
		case "gce":
			if network {
				gceCloudConfigFile, err := GetAndCreateGceDataSourceFilename()
				if err != nil {
					log.Errorf("Could not retrieve GCE CloudConfig %s", err)
					continue
				}
				dss = append(dss, file.NewDatasource(gceCloudConfigFile))
			}
		}
	}

	return dss
}

func enableDoLinkLocal() {
	err := netconf.ApplyNetworkConfigs(&netconf.NetworkConfig{
		Interfaces: map[string]netconf.InterfaceConfig{
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

func isCompose(content string) bool {
	return strings.HasPrefix(content, "#compose\n")
}

func composeToCloudConfig(bytes []byte) ([]byte, error) {
	compose := make(map[interface{}]interface{})
	err := yaml.Unmarshal(bytes, &compose)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(map[interface{}]interface{}{
		"rancher": map[interface{}]interface{}{
			"services": compose,
		},
	})
}
