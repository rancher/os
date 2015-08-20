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

	"gopkg.in/yaml.v2"

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
	"github.com/rancher/netconf"
	"github.com/rancherio/os/cmd/cloudinit/hostname"
	rancherConfig "github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

const (
	datasourceInterval    = 100 * time.Millisecond
	datasourceMaxInterval = 30 * time.Second
	datasourceTimeout     = 5 * time.Minute
	sshKeyName            = "rancheros-cloud-config"
	baseConfigDir         = "/var/lib/rancher/conf/cloud-config.d"
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
	os.Remove(rancherConfig.CloudConfigScriptFile)
	os.Remove(rancherConfig.CloudConfigFile)
	os.Remove(rancherConfig.MetaDataFile)

	if len(scriptBytes) > 0 {
		log.Infof("Writing to %s", rancherConfig.CloudConfigScriptFile)
		if err := ioutil.WriteFile(rancherConfig.CloudConfigScriptFile, scriptBytes, 500); err != nil {
			log.Errorf("Error while writing file %s: %v", rancherConfig.CloudConfigScriptFile, err)
			return err
		}
	}

	if err := ioutil.WriteFile(rancherConfig.CloudConfigFile, cloudConfigBytes, 400); err != nil {
		return err
	}
	log.Infof("Written to %s:\n%s", rancherConfig.CloudConfigFile, string(cloudConfigBytes))

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
		log.Fatalf("Failed to read rancher config %v", err)
	}

	dss := getDatasources(cfg)
	if len(dss) == 0 {
		return nil, nil
	}

	ds := selectDatasource(dss)
	return ds, nil
}

func mergeBaseConfig(current, currentScript []byte) ([]byte, []byte, error) {
	files, err := ioutil.ReadDir(baseConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Infof("%s does not exist, not merging", baseConfigDir)
			return current, currentScript, nil
		}

		log.Errorf("Failed to read %s: %v", baseConfigDir, err)
		return nil, nil, err
	}

	scriptResult := currentScript
	result := []byte{}

	for _, file := range files {
		if file.IsDir() || strings.HasPrefix(file.Name(), ".") {
			continue
		}

		input := path.Join(baseConfigDir, file.Name())
		content, err := ioutil.ReadFile(input)
		if err != nil {
			log.Errorf("Failed to read %s: %v", input, err)
			// ignore error
			continue
		}

		if config.IsScript(string(content)) {
			scriptResult = content
			continue
		}

		log.Infof("Merging %s", input)

		if isCompose(string(content)) {
			content, err = toCompose(content)
			if err != nil {
				log.Errorf("Failed to convert %s to cloud-config syntax: %v", input, err)
			}
		}

		result, err = util.MergeBytes(result, content)
		if err != nil {
			log.Errorf("Failed to merge bytes: %v", err)
			return nil, nil, err
		}
	}

	if len(result) == 0 {
		return current, scriptResult, nil
	} else {
		result, err := util.MergeBytes(result, current)
		return result, scriptResult, err
	}
}

func saveCloudConfig() error {
	var userDataBytes []byte
	var metadata datasource.Metadata

	ds, err := currentDatasource()
	if err != nil {
		log.Errorf("Failed to select datasource: %v", err)
		return err
	}

	if ds != nil {
		log.Infof("Fetching user-data from datasource %v", ds.Type())
		userDataBytes, err = ds.FetchUserdata()
		if err != nil {
			log.Errorf("Failed fetching user-data from datasource: %v", err)
			return err
		}

		log.Infof("Fetching meta-data from datasource of type %v", ds.Type())
		metadata, err = ds.FetchMetadata()
		if err != nil {
			log.Errorf("Failed fetching meta-data from datasource: %v", err)
			return err
		}
	}

	userDataBytes = substituteUserDataVars(userDataBytes, metadata)
	userData := string(userDataBytes)
	scriptBytes := []byte{}

	if config.IsScript(userData) {
		scriptBytes = userDataBytes
		userDataBytes = []byte{}
	} else if isCompose(userData) {
		if userDataBytes, err = toCompose(userDataBytes); err != nil {
			log.Errorf("Failed to convert to compose syntax: %v", err)
			return err
		}
	} else if config.IsCloudConfig(userData) {
		if rancherConfig.ReadConfig(userDataBytes) == nil {
			log.WithFields(log.Fields{"cloud-config": userData}).Warn("Failed to parse cloud-config, not saving.")
			userDataBytes = []byte{}
		}
	} else {
		log.Errorf("Unrecognized cloud-init\n%s", userData)
		userDataBytes = []byte{}
	}

	userDataBytesMerged, scriptBytes, err := mergeBaseConfig(userDataBytes, scriptBytes)
	if err != nil {
		log.Errorf("Failed to merge base config: %v", err)
	} else if rancherConfig.ReadConfig(userDataBytesMerged) == nil {
		log.WithFields(log.Fields{"cloud-config": userData}).Warn("Failed to parse merged cloud-config, not merging.")
	} else {
		userDataBytes = userDataBytesMerged
	}

	return saveFiles(userDataBytes, scriptBytes, metadata)
}

func getSaveCloudConfig() (*config.CloudConfig, error) {
	ds := file.NewDatasource(rancherConfig.CloudConfigFile)
	if !ds.IsAvailable() {
		log.Infof("%s does not exist", rancherConfig.CloudConfigFile)
		return nil, nil
	}

	ccBytes, err := ds.FetchUserdata()
	if err != nil {
		log.Errorf("Failed to read user-data from %s: %v", rancherConfig.CloudConfigFile, err)
		return nil, err
	}

	var cc config.CloudConfig
	err = yaml.Unmarshal(ccBytes, &cc)
	if err != nil {
		log.Errorf("Failed to unmarshall user-data from %s: %v", rancherConfig.CloudConfigFile, err)
		return nil, err
	}

	return &cc, err
}

func executeCloudConfig() error {
	ccu, err := getSaveCloudConfig()
	if err != nil {
		return err
	}

	var metadata datasource.Metadata

	metaDataBytes, err := ioutil.ReadFile(rancherConfig.MetaDataFile)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(metaDataBytes, &metadata); err != nil {
		return err
	}

	log.Info("Merging cloud-config from meta-data and user-data")
	cc := mergeConfigs(ccu, metadata)

	if cc.Hostname != "" {
		//set hostname
		if err := hostname.SetHostname(cc.Hostname); err != nil {
			log.Fatal(err)
		}
	}

	if len(cc.SSHAuthorizedKeys) > 0 {
		authorizeSSHKeys("rancher", cc.SSHAuthorizedKeys, sshKeyName)
		authorizeSSHKeys("docker", cc.SSHAuthorizedKeys, sshKeyName)
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
		fullPath, err := system.WriteFile(&f, "/")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
	}

	return nil
}

func Main() {
	flags.Parse(rancherConfig.FilterGlobalConfig(os.Args[1:]))

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

func toCompose(bytes []byte) ([]byte, error) {
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

func substituteUserDataVars(userDataBytes []byte, metadata datasource.Metadata) []byte {
	env := initialize.NewEnvironment("", "", "", "", metadata)
	userData := env.Apply(string(userDataBytes))

	return []byte(userData)
}
