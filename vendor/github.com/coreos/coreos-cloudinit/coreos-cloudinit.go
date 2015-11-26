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

package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/coreos/coreos-cloudinit/config"
	"github.com/coreos/coreos-cloudinit/config/validate"
	"github.com/coreos/coreos-cloudinit/datasource"
	"github.com/coreos/coreos-cloudinit/datasource/configdrive"
	"github.com/coreos/coreos-cloudinit/datasource/file"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/cloudsigma"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/digitalocean"
	"github.com/coreos/coreos-cloudinit/datasource/metadata/ec2"
	"github.com/coreos/coreos-cloudinit/datasource/proc_cmdline"
	"github.com/coreos/coreos-cloudinit/datasource/url"
	"github.com/coreos/coreos-cloudinit/datasource/waagent"
	"github.com/coreos/coreos-cloudinit/initialize"
	"github.com/coreos/coreos-cloudinit/network"
	"github.com/coreos/coreos-cloudinit/pkg"
	"github.com/coreos/coreos-cloudinit/system"
)

const (
	version               = "1.3.2+git"
	datasourceInterval    = 100 * time.Millisecond
	datasourceMaxInterval = 30 * time.Second
	datasourceTimeout     = 5 * time.Minute
)

var (
	flags = struct {
		printVersion  bool
		ignoreFailure bool
		sources       struct {
			file                        string
			configDrive                 string
			waagent                     string
			metadataService             bool
			ec2MetadataService          string
			cloudSigmaMetadataService   bool
			digitalOceanMetadataService string
			url                         string
			procCmdLine                 bool
		}
		convertNetconf string
		workspace      string
		sshKeyName     string
		oem            string
		validate       bool
	}{}
)

func init() {
	flag.BoolVar(&flags.printVersion, "version", false, "Print the version and exit")
	flag.BoolVar(&flags.ignoreFailure, "ignore-failure", false, "Exits with 0 status in the event of malformed input from user-data")
	flag.StringVar(&flags.sources.file, "from-file", "", "Read user-data from provided file")
	flag.StringVar(&flags.sources.configDrive, "from-configdrive", "", "Read data from provided cloud-drive directory")
	flag.StringVar(&flags.sources.waagent, "from-waagent", "", "Read data from provided waagent directory")
	flag.BoolVar(&flags.sources.metadataService, "from-metadata-service", false, "[DEPRECATED - Use -from-ec2-metadata] Download data from metadata service")
	flag.StringVar(&flags.sources.ec2MetadataService, "from-ec2-metadata", "", "Download EC2 data from the provided url")
	flag.BoolVar(&flags.sources.cloudSigmaMetadataService, "from-cloudsigma-metadata", false, "Download data from CloudSigma server context")
	flag.StringVar(&flags.sources.digitalOceanMetadataService, "from-digitalocean-metadata", "", "Download DigitalOcean data from the provided url")
	flag.StringVar(&flags.sources.url, "from-url", "", "Download user-data from provided url")
	flag.BoolVar(&flags.sources.procCmdLine, "from-proc-cmdline", false, fmt.Sprintf("Parse %s for '%s=<url>', using the cloud-config served by an HTTP GET to <url>", proc_cmdline.ProcCmdlineLocation, proc_cmdline.ProcCmdlineCloudConfigFlag))
	flag.StringVar(&flags.oem, "oem", "", "Use the settings specific to the provided OEM")
	flag.StringVar(&flags.convertNetconf, "convert-netconf", "", "Read the network config provided in cloud-drive and translate it from the specified format into networkd unit files")
	flag.StringVar(&flags.workspace, "workspace", "/var/lib/coreos-cloudinit", "Base directory coreos-cloudinit should use to store data")
	flag.StringVar(&flags.sshKeyName, "ssh-key-name", initialize.DefaultSSHKeyName, "Add SSH keys to the system with the given name")
	flag.BoolVar(&flags.validate, "validate", false, "[EXPERIMENTAL] Validate the user-data but do not apply it to the system")
}

type oemConfig map[string]string

var (
	oemConfigs = map[string]oemConfig{
		"digitalocean": oemConfig{
			"from-digitalocean-metadata": "http://169.254.169.254/",
			"convert-netconf":            "digitalocean",
		},
		"ec2-compat": oemConfig{
			"from-ec2-metadata": "http://169.254.169.254/",
			"from-configdrive":  "/media/configdrive",
		},
		"rackspace-onmetal": oemConfig{
			"from-configdrive": "/media/configdrive",
			"convert-netconf":  "debian",
		},
		"azure": oemConfig{
			"from-waagent": "/var/lib/waagent",
		},
		"cloudsigma": oemConfig{
			"from-cloudsigma-metadata": "true",
		},
	}
)

func main() {
	failure := false

	flag.Parse()

	if c, ok := oemConfigs[flags.oem]; ok {
		for k, v := range c {
			flag.Set(k, v)
		}
	} else if flags.oem != "" {
		oems := make([]string, 0, len(oemConfigs))
		for k := range oemConfigs {
			oems = append(oems, k)
		}
		fmt.Printf("Invalid option to --oem: %q. Supported options: %q\n", flags.oem, oems)
		os.Exit(2)
	}

	if flags.printVersion == true {
		fmt.Printf("coreos-cloudinit version %s\n", version)
		os.Exit(0)
	}

	switch flags.convertNetconf {
	case "":
	case "debian":
	case "digitalocean":
	default:
		fmt.Printf("Invalid option to -convert-netconf: '%s'. Supported options: 'debian, digitalocean'\n", flags.convertNetconf)
		os.Exit(2)
	}

	dss := getDatasources()
	if len(dss) == 0 {
		fmt.Println("Provide at least one of --from-file, --from-configdrive, --from-ec2-metadata, --from-cloudsigma-metadata, --from-url or --from-proc-cmdline")
		os.Exit(2)
	}

	ds := selectDatasource(dss)
	if ds == nil {
		fmt.Println("No datasources available in time")
		os.Exit(1)
	}

	fmt.Printf("Fetching user-data from datasource of type %q\n", ds.Type())
	userdataBytes, err := ds.FetchUserdata()
	if err != nil {
		fmt.Printf("Failed fetching user-data from datasource: %v\nContinuing...\n", err)
		failure = true
	}

	if report, err := validate.Validate(userdataBytes); err == nil {
		ret := 0
		for _, e := range report.Entries() {
			fmt.Println(e)
			ret = 1
		}
		if flags.validate {
			os.Exit(ret)
		}
	} else {
		fmt.Printf("Failed while validating user_data (%q)\n", err)
		if flags.validate {
			os.Exit(1)
		}
	}

	fmt.Printf("Fetching meta-data from datasource of type %q\n", ds.Type())
	metadata, err := ds.FetchMetadata()
	if err != nil {
		fmt.Printf("Failed fetching meta-data from datasource: %v\n", err)
		os.Exit(1)
	}

	// Apply environment to user-data
	env := initialize.NewEnvironment("/", ds.ConfigRoot(), flags.workspace, flags.sshKeyName, metadata)
	userdata := env.Apply(string(userdataBytes))

	var ccu *config.CloudConfig
	var script *config.Script
	if ud, err := initialize.ParseUserData(userdata); err != nil {
		fmt.Printf("Failed to parse user-data: %v\nContinuing...\n", err)
		failure = true
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

	var ifaces []network.InterfaceGenerator
	if flags.convertNetconf != "" {
		var err error
		switch flags.convertNetconf {
		case "debian":
			ifaces, err = network.ProcessDebianNetconf(metadata.NetworkConfig)
		case "digitalocean":
			ifaces, err = network.ProcessDigitalOceanNetconf(metadata.NetworkConfig)
		default:
			err = fmt.Errorf("Unsupported network config format %q", flags.convertNetconf)
		}
		if err != nil {
			fmt.Printf("Failed to generate interfaces: %v\n", err)
			os.Exit(1)
		}
	}

	if err = initialize.Apply(cc, ifaces, env); err != nil {
		fmt.Printf("Failed to apply cloud-config: %v\n", err)
		os.Exit(1)
	}

	if script != nil {
		if err = runScript(*script, env); err != nil {
			fmt.Printf("Failed to run script: %v\n", err)
			os.Exit(1)
		}
	}

	if failure && !flags.ignoreFailure {
		os.Exit(1)
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
func getDatasources() []datasource.Datasource {
	dss := make([]datasource.Datasource, 0, 5)
	if flags.sources.file != "" {
		dss = append(dss, file.NewDatasource(flags.sources.file))
	}
	if flags.sources.url != "" {
		dss = append(dss, url.NewDatasource(flags.sources.url))
	}
	if flags.sources.configDrive != "" {
		dss = append(dss, configdrive.NewDatasource(flags.sources.configDrive))
	}
	if flags.sources.metadataService {
		dss = append(dss, ec2.NewDatasource(ec2.DefaultAddress))
	}
	if flags.sources.ec2MetadataService != "" {
		dss = append(dss, ec2.NewDatasource(flags.sources.ec2MetadataService))
	}
	if flags.sources.cloudSigmaMetadataService {
		dss = append(dss, cloudsigma.NewServerContextService())
	}
	if flags.sources.digitalOceanMetadataService != "" {
		dss = append(dss, digitalocean.NewDatasource(flags.sources.digitalOceanMetadataService))
	}
	if flags.sources.waagent != "" {
		dss = append(dss, waagent.NewDatasource(flags.sources.waagent))
	}
	if flags.sources.procCmdLine {
		dss = append(dss, proc_cmdline.NewDatasource())
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
