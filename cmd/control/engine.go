package control

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/rancher/os/cmd/control/service"
	"github.com/rancher/os/cmd/control/service/app"
	"github.com/rancher/os/config"
	"github.com/rancher/os/pkg/compose"
	"github.com/rancher/os/pkg/docker"
	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"
	"github.com/rancher/os/pkg/util/network"

	yaml "github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/reference"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	composeConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project/options"
	composeYaml "github.com/docker/libcompose/yaml"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

func engineSubcommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "switch",
			Usage:  "switch user Docker engine without a reboot",
			Action: engineSwitch,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "do not prompt for input",
				},
				cli.BoolFlag{
					Name:  "no-pull",
					Usage: "don't pull console image",
				},
			},
		},
		{
			Name:        "create",
			Usage:       "create Dind engine without a reboot",
			Description: "must switch user docker to 17.12.1 or earlier if using Dind",
			ArgsUsage:   "<name>",
			Before:      preFlightValidate,
			Action:      engineCreate,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "version, v",
					Value: config.DefaultDind,
					Usage: fmt.Sprintf("set the version for the engine, %s are available", config.SupportedDinds),
				},
				cli.StringFlag{
					Name:  "network",
					Usage: "set the network for the engine",
				},
				cli.StringFlag{
					Name:  "fixed-ip",
					Usage: "set the fixed ip for the engine",
				},
				cli.StringFlag{
					Name:  "ssh-port",
					Usage: "set the ssh port for the engine",
				},
				cli.StringFlag{
					Name:  "authorized-keys",
					Usage: "set the ssh authorized_keys absolute path for the engine",
				},
			},
		},
		{
			Name:      "rm",
			Usage:     "remove Dind engine without a reboot",
			ArgsUsage: "<name>",
			Before: func(c *cli.Context) error {
				if len(c.Args()) != 1 {
					return errors.New("Must specify exactly one Docker engine to remove")
				}
				return nil
			},
			Action: dindEngineRemove,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "timeout,t",
					Usage: "specify a shutdown timeout in seconds",
					Value: 10,
				},
				cli.BoolFlag{
					Name:  "force, f",
					Usage: "do not prompt for input",
				},
			},
		},
		{
			Name:   "enable",
			Usage:  "set user Docker engine to be switched on next reboot",
			Action: engineEnable,
		},
		{
			Name:   "list",
			Usage:  "list available Docker engines (include the Dind engines)",
			Action: engineList,
		},
	}
}

func engineSwitch(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one Docker engine to switch to")
	}
	newEngine := c.Args()[0]

	cfg := config.LoadConfig()
	validateEngine(newEngine, cfg)

	project, err := compose.GetProject(cfg, true, false)
	if err != nil {
		log.Fatal(err)
	}

	if err = project.Stop(context.Background(), 10, "docker"); err != nil {
		log.Fatal(err)
	}

	if err = compose.LoadSpecialService(project, cfg, "docker", newEngine); err != nil {
		log.Fatal(err)
	}

	if err = project.Up(context.Background(), options.Up{}, "docker"); err != nil {
		log.Fatal(err)
	}

	if err := config.Set("rancher.docker.engine", newEngine); err != nil {
		log.Errorf("Failed to update rancher.docker.engine: %v", err)
	}

	return nil
}

func engineCreate(c *cli.Context) error {
	name := c.Args()[0]
	version := c.String("version")
	sshPort, _ := strconv.Atoi(c.String("ssh-port"))
	if sshPort <= 0 {
		sshPort = randomSSHPort()
	}
	authorizedKeys := c.String("authorized-keys")
	network := c.String("network")
	fixedIP := c.String("fixed-ip")

	// generate & create engine compose
	err := generateEngineCompose(name, version, sshPort, authorizedKeys, network, fixedIP)
	if err != nil {
		return err
	}

	// stage engine service
	cfg := config.LoadConfig()
	var enabledServices []string
	if val, ok := cfg.Rancher.ServicesInclude[name]; !ok || !val {
		cfg.Rancher.ServicesInclude[name] = true
		enabledServices = append(enabledServices, name)
	}

	if len(enabledServices) > 0 {
		if err := compose.StageServices(cfg, enabledServices...); err != nil {
			log.Fatal(err)
		}

		if err := config.Set("rancher.services_include", cfg.Rancher.ServicesInclude); err != nil {
			log.Fatal(err)
		}
	}

	// generate engine script
	err = util.GenerateDindEngineScript(name)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func dindEngineRemove(c *cli.Context) error {
	if !c.Bool("force") {
		if !yes("Continue") {
			return nil
		}
	}

	// app.ProjectDelete needs to use this flag
	// Allow deletion of the Dind engine
	c.Set("force", "true")
	// Remove volumes associated with the Dind engine container
	c.Set("v", "true")

	name := c.Args()[0]
	cfg := config.LoadConfig()
	p, err := compose.GetProject(cfg, true, false)
	if err != nil {
		log.Fatalf("Get project failed: %v", err)
	}

	// 1. service stop
	err = app.ProjectStop(p, c)
	if err != nil {
		log.Fatalf("Stop project service failed: %v", err)
	}

	// 2. service delete
	err = app.ProjectDelete(p, c)
	if err != nil {
		log.Fatalf("Delete project service failed: %v", err)
	}

	// 3. service delete
	if _, ok := cfg.Rancher.ServicesInclude[name]; !ok {
		log.Fatalf("Failed to found enabled service %s", name)
	}

	delete(cfg.Rancher.ServicesInclude, name)

	if err = config.Set("rancher.services_include", cfg.Rancher.ServicesInclude); err != nil {
		log.Fatal(err)
	}

	// 4. remove service from file
	err = RemoveEngineFromCompose(name)
	if err != nil {
		log.Fatal(err)
	}

	// 5. remove dind engine script
	err = util.RemoveDindEngineScript(name)
	if err != nil {
		return err
	}

	return nil
}

func engineEnable(c *cli.Context) error {
	if len(c.Args()) != 1 {
		log.Fatal("Must specify exactly one Docker engine to enable")
	}
	newEngine := c.Args()[0]

	cfg := config.LoadConfig()
	validateEngine(newEngine, cfg)

	if err := compose.StageServices(cfg, newEngine); err != nil {
		return err
	}

	if err := config.Set("rancher.docker.engine", newEngine); err != nil {
		log.Errorf("Failed to update 'rancher.docker.engine': %v", err)
	}

	return nil
}

func engineList(c *cli.Context) error {
	cfg := config.LoadConfig()
	engines := availableEngines(cfg)
	currentEngine := CurrentEngine()

	for _, engine := range engines {
		if engine == currentEngine {
			fmt.Printf("current  %s\n", engine)
		} else if engine == cfg.Rancher.Docker.Engine {
			fmt.Printf("enabled  %s\n", engine)
		} else {
			fmt.Printf("disabled %s\n", engine)
		}
	}

	// check the dind container
	client, err := docker.NewSystemClient()
	if err != nil {
		log.Warnf("Failed to detect dind: %v", err)
		return nil
	}

	filter := filters.NewArgs()
	filter.Add("label", config.UserDockerLabel)
	opts := types.ContainerListOptions{
		All:    true,
		Filter: filter,
	}
	containers, err := client.ContainerList(context.Background(), opts)
	if err != nil {
		log.Warnf("Failed to detect dind: %v", err)
		return nil
	}
	for _, c := range containers {
		if c.State == "running" {
			fmt.Printf("enabled  %s\n", c.Labels[config.UserDockerLabel])
		} else {
			fmt.Printf("disabled %s\n", c.Labels[config.UserDockerLabel])
		}
	}

	return nil
}

func validateEngine(engine string, cfg *config.CloudConfig) {
	engines := availableEngines(cfg)
	if !service.IsLocalOrURL(engine) && !util.Contains(engines, engine) {
		log.Fatalf("%s is not a valid engine", engine)
	}
}

func availableEngines(cfg *config.CloudConfig) []string {
	engines, err := network.GetEngines(cfg.Rancher.Repositories.ToArray())
	if err != nil {
		log.Fatal(err)
	}
	sort.Strings(engines)
	return engines
}

// CurrentEngine gets the name of the docker that's running
func CurrentEngine() (engine string) {
	// sudo system-docker inspect --format "{{.Config.Image}}" docker
	client, err := docker.NewSystemClient()
	if err != nil {
		log.Warnf("Failed to detect current docker: %v", err)
		return
	}
	info, err := client.ContainerInspect(context.Background(), "docker")
	if err != nil {
		log.Warnf("Failed to detect current docker: %v", err)
		return
	}
	// parse image name, then remove os- prefix and the engine suffix
	image, err := reference.ParseNamed(info.Config.Image)
	if err != nil {
		log.Warnf("Failed to detect current docker(%s): %v", info.Config.Image, err)
		return
	}
	if t, ok := image.(reference.NamedTagged); ok {
		tag := t.Tag()
		if !strings.HasPrefix(tag, "1.") {
			// TODO: this assumes we only do Docker ce :/
			tag = tag + "-ce"
		}
		return "docker-" + tag
	}

	return
}

func preFlightValidate(c *cli.Context) error {
	if len(c.Args()) != 1 {
		return errors.New("Must specify one engine name")
	}
	name := c.Args()[0]
	if name == "" {
		return errors.New("Must specify one engine name")
	}

	version := c.String("version")
	if version == "" {
		return errors.New("Must specify one engine version")
	}

	authorizedKeys := c.String("authorized-keys")
	if authorizedKeys != "" {
		if _, err := os.Stat(authorizedKeys); os.IsNotExist(err) {
			return errors.New("The authorized-keys should be an exist file, recommended to put in the /opt or /var/lib/rancher directory")
		}
	}

	network := c.String("network")
	if network == "" {
		return errors.New("Must specify network")
	}

	userDefineNetwork, err := CheckUserDefineNetwork(network)
	if err != nil {
		return err
	}

	fixedIP := c.String("fixed-ip")
	if fixedIP == "" {
		return errors.New("Must specify fix ip")
	}

	err = CheckUserDefineIPv4Address(fixedIP, *userDefineNetwork)
	if err != nil {
		return err
	}

	isVersionMatch := false
	for _, v := range config.SupportedDinds {
		if v == version {
			isVersionMatch = true
			break
		}
	}

	if !isVersionMatch {
		return errors.Errorf("Engine version not supported only %v are supported", config.SupportedDinds)
	}

	if c.String("ssh-port") != "" {
		port, err := strconv.Atoi(c.String("ssh-port"))
		if err != nil {
			return errors.Wrap(err, "Failed to convert ssh port to Int")
		}
		if port > 0 {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:"+strconv.Itoa(port))
			if err != nil {
				return errors.Errorf("Failed to resolve tcp addr: %v", err)
			}
			l, err := net.ListenTCP("tcp", addr)
			if err != nil {
				return errors.Errorf("Failed to listen tcp: %v", err)
			}
			defer l.Close()
		}
	}

	return nil
}

func randomSSHPort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Errorf("Failed to resolve tcp addr: %v", err)
		return 0
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func generateEngineCompose(name, version string, sshPort int, authorizedKeys, network, fixedIP string) error {
	if err := os.MkdirAll(path.Dir(config.MultiDockerConfFile), 0700); err != nil && !os.IsExist(err) {
		log.Errorf("Failed to create directory for file %s: %v", config.MultiDockerConfFile, err)
		return err
	}

	composeConfigs := map[string]composeConfig.ServiceConfigV1{}

	if _, err := os.Stat(config.MultiDockerConfFile); err == nil {
		// read from engine compose
		bytes, err := ioutil.ReadFile(config.MultiDockerConfFile)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(bytes, &composeConfigs)
		if err != nil {
			return err
		}
	}

	if err := os.MkdirAll(config.MultiDockerDataDir+"/"+name, 0700); err != nil && !os.IsExist(err) {
		log.Errorf("Failed to create directory for file %s: %v", config.MultiDockerDataDir+"/"+name, err)
		return err
	}

	volumes := []string{
		"/lib/modules:/lib/modules",
		config.MultiDockerDataDir + "/" + name + ":" + config.MultiDockerDataDir + "/" + name,
	}
	if authorizedKeys != "" {
		volumes = append(volumes, authorizedKeys+":/root/.ssh/authorized_keys")
	}

	composeConfigs[name] = composeConfig.ServiceConfigV1{
		Image:       "${REGISTRY_DOMAIN}/" + version,
		Restart:     "always",
		Privileged:  true,
		Net:         network,
		Ports:       []string{strconv.Itoa(sshPort) + ":22"},
		Volumes:     volumes,
		VolumesFrom: []string{},
		Command: composeYaml.Command{
			"--storage-driver=overlay2",
			"--data-root=" + config.MultiDockerDataDir + "/" + name,
			"--host=unix://" + config.MultiDockerDataDir + "/" + name + "/docker-" + name + ".sock",
		},
		Labels: composeYaml.SliceorMap{
			"io.rancher.os.scope":     "system",
			"io.rancher.os.after":     "console",
			config.UserDockerLabel:    name,
			config.UserDockerNetLabel: network,
			config.UserDockerFIPLabel: fixedIP,
		},
	}

	bytes, err := yaml.Marshal(composeConfigs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(config.MultiDockerConfFile, bytes, 0640)
}

func RemoveEngineFromCompose(name string) error {
	composeConfigs := map[string]composeConfig.ServiceConfigV1{}

	if _, err := os.Stat(config.MultiDockerConfFile); err == nil {
		// read from engine compose
		bytes, err := ioutil.ReadFile(config.MultiDockerConfFile)
		if err != nil {
			return err
		}
		err = yaml.Unmarshal(bytes, &composeConfigs)
		if err != nil {
			return err
		}
	}

	delete(composeConfigs, name)

	bytes, err := yaml.Marshal(composeConfigs)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(config.MultiDockerConfFile, bytes, 0640)
}

func CheckUserDefineNetwork(name string) (*types.NetworkResource, error) {
	systemClient, err := docker.NewSystemClient()
	if err != nil {
		return nil, err
	}

	networks, err := systemClient.NetworkList(context.Background(), types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}

	for _, network := range networks {
		if network.Name == name {
			return &network, nil
		}
	}

	return nil, errors.Errorf("Failed to found the user define network: %s", name)
}

func CheckUserDefineIPv4Address(ipv4 string, network types.NetworkResource) error {
	for _, config := range network.IPAM.Config {
		_, ipnet, _ := net.ParseCIDR(config.Subnet)
		if ipnet.Contains(net.ParseIP(ipv4)) {
			return nil
		}
	}
	return errors.Errorf("IP %s is not in the specified cidr", ipv4)
}
