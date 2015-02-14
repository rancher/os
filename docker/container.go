package docker

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/runconfig"
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

const (
	LABEL = "label"
	HASH  = "io.rancher.os.hash"
)

type Container struct {
	Err          error
	Name         string
	remove       bool
	detach       bool
	Config       *runconfig.Config
	HostConfig   *runconfig.HostConfig
	cfg          *config.Config
	container    *dockerClient.Container
	containerCfg *config.ContainerConfig
}

func getHash(containerCfg *config.ContainerConfig) (string, error) {
	hash := sha1.New()
	w := util.NewErrorWriter(hash)

	w.Write([]byte(containerCfg.Id))
	w.Write([]byte(strings.Join(containerCfg.Cmd, ":")))

	if w.Err != nil {
		return "", w.Err
	}

	return hex.EncodeToString(hash.Sum([]byte{})), nil
}

func StartAndWait(cfg *config.Config, containerCfg *config.ContainerConfig) error {
	container := NewContainer(cfg, containerCfg).start(true)
	return container.Err
}

func NewContainer(cfg *config.Config, containerCfg *config.ContainerConfig) *Container {
	return &Container{
		cfg:          cfg,
		containerCfg: containerCfg,
	}
}

func (c *Container) returnErr(err error) *Container {
	c.Err = err
	return c
}

func (c *Container) Lookup() *Container {
	c.Parse()

	if c.Err != nil || (c.container != nil && c.container.HostConfig != nil) {
		return c
	}

	hash, err := getHash(c.containerCfg)
	if err != nil {
		return c.returnErr(err)
	}

	client, err := NewClient(c.cfg)
	if err != nil {
		return c.returnErr(err)
	}

	containers, err := client.ListContainers(dockerClient.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			LABEL: []string{fmt.Sprintf("%s=%s", HASH, hash)},
		},
	})
	if err != nil {
		return c.returnErr(err)
	}

	if len(containers) == 0 {
		return c
	}

	c.container, c.Err = client.InspectContainer(containers[0].ID)

	return c
}

func (c *Container) Exists() bool {
	c.Lookup()
	return c.container != nil
}

func (c *Container) Reset() *Container {
	c.Config = nil
	c.HostConfig = nil
	c.container = nil
	c.Err = nil

	return c
}

func (c *Container) Parse() *Container {
	if c.Config != nil || c.Err != nil {
		return c
	}

	flags := flag.NewFlagSet("run", flag.ExitOnError)

	flRemove := flags.Bool([]string{"#rm", "-rm"}, false, "")
	flDetach := flags.Bool([]string{"d", "-detach"}, false, "")
	flName := flags.String([]string{"#name", "-name"}, "", "")

	c.Config, c.HostConfig, _, c.Err = runconfig.Parse(flags, c.containerCfg.Cmd)

	c.Name = *flName
	c.detach = *flDetach
	c.remove = *flRemove

	if len(c.containerCfg.Id) == 0 {
		c.containerCfg.Id = c.Name
	}

	return c
}

func (c *Container) Start() *Container {
	return c.start(false)
}

func (c *Container) Stage() *Container {
	c.Parse()

	if c.Err != nil {
		return c
	}

	client, err := NewClient(c.cfg)
	if err != nil {
		c.Err = err
		return c
	}

	_, err = client.InspectImage(c.Config.Image)
	if err == dockerClient.ErrNoSuchImage {
		c.Err = client.PullImage(dockerClient.PullImageOptions{
			Repository:   c.Config.Image,
			OutputStream: os.Stdout,
		}, dockerClient.AuthConfiguration{})
	} else if err != nil {
		c.Err = err
	}

	return c
}

func (c *Container) Delete() *Container {
	c.Parse()
	c.Stage()
	c.Lookup()

	if c.Err != nil {
		return c
	}

	if !c.Exists() {
		return c
	}

	client, err := NewClient(c.cfg)
	if err != nil {
		return c.returnErr(err)
	}

	err = client.RemoveContainer(dockerClient.RemoveContainerOptions{
		ID:    c.container.ID,
		Force: true,
	})
	if err != nil {
		return c.returnErr(err)
	}

	return c
}

func renameOld(client *dockerClient.Client, opts *dockerClient.CreateContainerOptions) error {
	if len(opts.Name) == 0 {
		return nil
	}

	existing, err := client.InspectContainer(opts.Name)
	if _, ok := err.(dockerClient.NoSuchContainer); ok {
		return nil
	}
	if err != nil {
		return nil
	}

	if label, ok := existing.Config.Labels[HASH]; ok {
		return client.RenameContainer(existing.ID, fmt.Sprintf("%s-%s", existing.Name, label))
	} else {
		//TODO: do something with containers with no hash
		return errors.New("Existing container doesn't have a hash")
	}
}

func (c *Container) start(wait bool) *Container {
	c.Lookup()
	c.Stage()

	if c.Err != nil {
		return c
	}

	bytes, err := json.Marshal(c)
	if err != nil {
		return c.returnErr(err)
	}

	client, err := NewClient(c.cfg)
	if err != nil {
		return c.returnErr(err)
	}

	var opts dockerClient.CreateContainerOptions
	container := c.container
	created := false

	if !c.Exists() {
		c.Err = json.Unmarshal(bytes, &opts)
		if c.Err != nil {
			return c
		}

		if opts.Config.Labels == nil {
			opts.Config.Labels = make(map[string]string)
		}

		hash, err := getHash(c.containerCfg)
		if err != nil {
			return c.returnErr(err)
		}

		opts.Config.Labels[HASH] = hash

		err = renameOld(client, &opts)
		if err != nil {
			return c.returnErr(err)
		}

		container, err = client.CreateContainer(opts)
		created = true
		if err != nil {
			return c.returnErr(err)
		}
	}

	c.container = container

	hostConfig := container.HostConfig
	if created {
		hostConfig = opts.HostConfig
	}

	err = client.StartContainer(container.ID, hostConfig)
	if err != nil {
		return c.returnErr(err)
	}

	if !c.detach && wait {
		_, c.Err = client.WaitContainer(container.ID)
		return c
	}

	return c
}
