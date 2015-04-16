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
	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
	rancherNetwork "github.com/rancherio/os/cmd/network"
	rancherConfig "github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
	"gopkg.in/yaml.v2"
)

const (
	datasourceInterval    = 100 * time.Millisecond
	datasourceMaxInterval = 30 * time.Second
	datasourceTimeout     = 5 * time.Minute
)

var (
	baseConfigDir string
	outputDir     string
	outputFile    string
	metaDataFile  string
	scriptFile    string
	rancherYml    string
	save          bool
	execute       bool
	network       bool
	sshKeyName    string
	flags         *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.StringVar(&baseConfigDir, "base-config-dir", "/var/lib/rancher/conf/cloud-config.d", "base cloud config")
	flags.StringVar(&outputDir, "dir", "/var/lib/rancher/conf", "working directory")
	flags.StringVar(&outputFile, "file", "cloud-config-processed.yml", "output cloud config file name")
	flags.StringVar(&metaDataFile, "metadata", "metadata", "output metdata file name")
	flags.StringVar(&scriptFile, "script-file", "cloud-config-script", "output cloud config script file name")
	flags.StringVar(&rancherYml, "rancher", "cloud-config-rancher.yml", "output cloud config rancher file name")
	flags.StringVar(&sshKeyName, "ssh-key-name", "rancheros-cloud-config", "SSH key name")
	flags.BoolVar(&network, "network", true, "use network based datasources")
	flags.BoolVar(&save, "save", false, "save cloud config and exit")
	flags.BoolVar(&execute, "execute", false, "execute saved cloud config")
}

func saveFiles(cloudConfigBytes, scriptBytes []byte, metadata datasource.Metadata) error {
	scriptOutput := path.Join(outputDir, scriptFile)
	cloudConfigOutput := path.Join(outputDir, outputFile)
	rancherYmlOutput := path.Join(outputDir, rancherYml)
	metaDataOutput := path.Join(outputDir, metaDataFile)

	os.Remove(scriptOutput)
	os.Remove(cloudConfigOutput)
	os.Remove(rancherYmlOutput)
	os.Remove(metaDataOutput)

	if len(scriptBytes) > 0 {
		log.Infof("Writing to %s", scriptOutput)
		if err := ioutil.WriteFile(scriptOutput, scriptBytes, 500); err != nil {
			log.Errorf("Error while writing file %s: %v", scriptOutput, err)
			return err
		}
	}

	cloudConfigBytes = append([]byte("#cloud-config\n"), cloudConfigBytes...)
	log.Infof("Writing to %s", cloudConfigOutput)
	if err := ioutil.WriteFile(cloudConfigOutput, cloudConfigBytes, 500); err != nil {
		log.Errorf("Error while writing file %s: %v", cloudConfigOutput, err)
		return err
	}

	ccData := make(map[string]interface{})
	if err := yaml.Unmarshal(cloudConfigBytes, ccData); err != nil {
		return err
	}

	if rancher, ok := ccData["rancher"]; ok {
		bytes, err := yaml.Marshal(rancher)
		if err != nil {
			return err
		}

		if err = ioutil.WriteFile(rancherYmlOutput, bytes, 400); err != nil {
			return err
		}
	}

	metaDataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(metaDataOutput, metaDataBytes, 400); err != nil {
		return err
	}

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

func mergeBaseConfig(current []byte) ([]byte, error) {
	files, err := ioutil.ReadDir(baseConfigDir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Infof("%s does not exist, not merging", baseConfigDir)
			return current, nil
		}

		log.Errorf("Failed to read %s: %v", baseConfigDir, err)
		return nil, err
	}

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

		log.Infof("Merging %s", input)
		result, err = util.MergeBytes(result, content)
		if err != nil {
			log.Errorf("Failed to merge bytes: %v", err)
			return nil, err
		}
	}

	if len(result) == 0 {
		return current, nil
	} else {
		return util.MergeBytes(result, current)
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
		// nothing to do
	} else {
		log.Errorf("Unrecognized cloud-init\n%s", userData)
		userDataBytes = []byte{}
	}

	if userDataBytes, err = mergeBaseConfig(userDataBytes); err != nil {
		log.Errorf("Failed to merge base config: %v", err)
		return err
	}

	return saveFiles(userDataBytes, scriptBytes, metadata)
}

func getSaveCloudConfig() (*config.CloudConfig, error) {
	cloudConfig := path.Join(outputDir, outputFile)

	ds := file.NewDatasource(cloudConfig)
	if !ds.IsAvailable() {
		log.Infof("%s does not exist", cloudConfig)
		return nil, nil
	}

	ccBytes, err := ds.FetchUserdata()
	if err != nil {
		log.Errorf("Failed to read user-data from %s: %v", cloudConfig, err)
		return nil, err
	}

	var cc config.CloudConfig
	err = yaml.Unmarshal(ccBytes, &cc)
	if err != nil {
		log.Errorf("Failed to unmarshall user-data from %s: %v", cloudConfig, err)
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

	metaDataBytes, err := ioutil.ReadFile(path.Join(outputDir, metaDataFile))
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(metaDataBytes, &metadata); err != nil {
		return err
	}

	log.Info("Merging cloud-config from meta-data and user-data")
	cc := mergeConfigs(ccu, metadata)

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
		fullPath, err := system.WriteFile(&f, "/")
		if err != nil {
			log.Fatalf("%v", err)
		}
		log.Printf("Wrote file %s to filesystem", fullPath)
	}

	return nil
}

func Main() {
	flags.Parse(rancherConfig.FilterGlobalConfig(os.Args[1:]))

	if save {
		err := saveCloudConfig()
		if err != nil {
			log.Fatalf("Failed to save cloud config: %v", err)
		}
	}

	if execute {
		err := executeCloudConfig()
		if err != nil {
			log.Fatalf("Failed to save cloud config: %v", err)
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
func getDatasources(cfg *rancherConfig.Config) []datasource.Datasource {
	dss := make([]datasource.Datasource, 0, 5)

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

func isCompose(content string) bool {
	return strings.HasPrefix(content, "#compose\n")
}

func toCompose(bytes []byte) ([]byte, error) {
	result := make(map[interface{}]interface{})
	compose := make(map[interface{}]interface{})
	err := yaml.Unmarshal(bytes, &compose)
	if err != nil {
		return nil, err
	}

	result["services"] = compose
	return yaml.Marshal(result)
}
