package cloudinit

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
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
	flags      *flag.FlagSet
)

func init() {
	flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flags.StringVar(&outputDir, "dir", "/var/lib/rancher/conf", "working directory")
	flags.StringVar(&outputFile, "file", "cloud-config.yml", "cloud config file name")
	flags.StringVar(&sshKeyName, "ssh-key-name", "rancheros-cloud-config", "SSH key name")
	flags.BoolVar(&save, "save", false, "save cloud config and exit")
}

func Main() {
	flags.Parse(os.Args[1:])

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

	if report, err := validate.Validate(userdataBytes); err == nil {
		fail := false
		for _, e := range report.Entries() {
			log.Error(e)
			fail = true
		}
		if fail {
			fmt.Println("failed validation")
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

		output := path.Join(outputDir, outputFile)
		if err := ioutil.WriteFile(output, fileData, 400); err != nil {
			log.Fatalf("Error while writing file %v", err)
		}

		os.Exit(0)
	}

	if script != nil {
		if ds.Type() != "local-file" {
			fmt.Println("can only execute local files")
		}
		cmdPath := reflect.ValueOf(ds).Elem().Field(0).String()
		cmd := exec.Command(cmdPath)
		fmt.Println("running ", cmdPath)
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to run script: %v\n", err)
			os.Exit(1)
		}
	}

	if &cc == nil {
		log.Fatal("no config or script found")	
	}

	for _, user := range cc.Users {
		if user.Name == "" {
			continue
		}
		if len(user.SSHAuthorizedKeys) > 0 {
			authorizeSSHKeys(user.Name, user.SSHAuthorizedKeys, env.SSHKeyName())
		}
	}
	
	for _, file := range cc.WriteFiles {
		f := system.File{File: file}
		fullPath, err := system.WriteFile(&f, env.Root())
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
		parts := strings.SplitN(ds, ":", 2)

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

