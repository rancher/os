package docker

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/runconfig"
	shlex "github.com/flynn/go-shlex"
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/util"
)

const (
	LABEL = "label"
	HASH  = "io.rancher.os.hash"
	ID    = "io.rancher.os.id"
)

type Container struct {
	Err          error
	Name         string
	remove       bool
	detach       bool
	Config       *runconfig.Config
	HostConfig   *runconfig.HostConfig
	dockerHost   string
	Container    *dockerClient.Container
	ContainerCfg *config.ContainerConfig
}

type ByCreated []dockerClient.APIContainers

func (c ByCreated) Len() int           { return len(c) }
func (c ByCreated) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c ByCreated) Less(i, j int) bool { return c[j].Created < c[i].Created }

func getHash(containerCfg *config.ContainerConfig) (string, error) {
	hash := sha1.New()
	w := util.NewErrorWriter(hash)

	w.Write([]byte(containerCfg.Id))
	w.Write([]byte(containerCfg.Cmd))

	if w.Err != nil {
		return "", w.Err
	}

	return hex.EncodeToString(hash.Sum([]byte{})), nil
}

func StartAndWait(dockerHost string, containerCfg *config.ContainerConfig) error {
	container := NewContainer(dockerHost, containerCfg).start(false, true)
	return container.Err
}

func NewContainer(dockerHost string, containerCfg *config.ContainerConfig) *Container {
	c := &Container{
		dockerHost:   dockerHost,
		ContainerCfg: containerCfg,
	}
	return c.Parse()
}

func (c *Container) returnErr(err error) *Container {
	c.Err = err
	return c
}

func getByLabel(client *dockerClient.Client, key, value string) (*dockerClient.APIContainers, error) {
	containers, err := client.ListContainers(dockerClient.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			LABEL: []string{fmt.Sprintf("%s=%s", key, value)},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, nil
	}

	sort.Sort(ByCreated(containers))
	return &containers[0], nil
}

func (c *Container) Lookup() *Container {
	c.Parse()

	if c.Err != nil || (c.Container != nil && c.Container.HostConfig != nil) {
		return c
	}

	hash, err := getHash(c.ContainerCfg)
	if err != nil {
		return c.returnErr(err)
	}

	client, err := NewClient(c.dockerHost)
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

	c.Container, c.Err = inspect(client, containers[0].ID)

	return c
}

func (c *Container) Log(stdout io.Writer, stderr io.Writer, follow bool) error {
	client, err := NewClient(c.dockerHost)
	if err != nil {
		return err
	}

	return client.Logs(dockerClient.LogsOptions{
		Container:    c.Name,
		Stdout:       true,
		Stderr:       true,
		Follow:       follow,
		OutputStream: stdout,
		ErrorStream:  stderr,
	})
}

func inspect(client *dockerClient.Client, id string) (*dockerClient.Container, error) {
	c, err := client.InspectContainer(id)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(c.Name, "/") {
		c.Name = c.Name[1:]
	}

	return c, err
}

func (c *Container) Exists() bool {
	c.Lookup()
	return c.Container != nil
}

func (c *Container) Reset() *Container {
	c.Config = nil
	c.HostConfig = nil
	c.Container = nil
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

	args, err := shlex.Split(c.ContainerCfg.Cmd)
	if err != nil {
		return c.returnErr(err)
	}

	log.Debugf("Parsing [%s]", strings.Join(args, ","))
	c.Config, c.HostConfig, _, c.Err = runconfig.Parse(flags, args)

	c.Name = *flName
	c.detach = *flDetach
	c.remove = *flRemove

	if c.ContainerCfg.Id == "" {
		c.ContainerCfg.Id = c.Name
	}

	return c
}

func (c *Container) Create() *Container {
	return c.start(true, false)
}

func (c *Container) Start() *Container {
	return c.start(false, false)
}

func (c *Container) StartAndWait() *Container {
	return c.start(false, true)
}

func (c *Container) Stage() *Container {
	c.Parse()

	if c.Err != nil {
		return c
	}

	client, err := NewClient(c.dockerHost)
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

	client, err := NewClient(c.dockerHost)
	if err != nil {
		return c.returnErr(err)
	}

	err = client.RemoveContainer(dockerClient.RemoveContainerOptions{
		ID:    c.Container.ID,
		Force: true,
	})
	if err != nil {
		return c.returnErr(err)
	}

	return c
}

func (c *Container) renameCurrent(client *dockerClient.Client) error {
	if c.Name == "" {
		return nil
	}

	if c.Name == c.Container.Name {
		return nil
	}

	err := client.RenameContainer(dockerClient.RenameContainerOptions{ID: c.Container.ID, Name: c.Name})
	if err != nil {
		return err
	}

	c.Container, err = inspect(client, c.Container.ID)
	return err
}

func (c *Container) renameOld(client *dockerClient.Client, opts *dockerClient.CreateContainerOptions) error {
	if len(opts.Name) == 0 {
		return nil
	}

	existing, err := inspect(client, opts.Name)
	if _, ok := err.(*dockerClient.NoSuchContainer); ok {
		return nil
	}

	if err != nil {
		return nil
	}

	if c.Container != nil && existing.ID == c.Container.ID {
		return nil
	}

	var newName string
	if label, ok := existing.Config.Labels[HASH]; ok {
		newName = fmt.Sprintf("%s-%s", existing.Name, label)
	} else {
		newName = fmt.Sprintf("%s-unknown-%s", existing.Name, util.RandSeq(12))
	}

	if existing.State.Running {
		err := client.StopContainer(existing.ID, 2)
		if err != nil {
			return err
		}

		_, err = client.WaitContainer(existing.ID)
		if err != nil {
			return err
		}
	}

	log.Debugf("Renaming %s to %s", existing.Name, newName)
	return client.RenameContainer(dockerClient.RenameContainerOptions{ID: existing.ID, Name: newName})
}

func (c *Container) getCreateOpts(client *dockerClient.Client) (*dockerClient.CreateContainerOptions, error) {
	bytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	var opts dockerClient.CreateContainerOptions

	err = json.Unmarshal(bytes, &opts)
	if err != nil {
		return nil, err
	}

	if opts.Config.Labels == nil {
		opts.Config.Labels = make(map[string]string)
	}

	hash, err := getHash(c.ContainerCfg)
	if err != nil {
		return nil, err
	}

	opts.Config.Labels[HASH] = hash
	opts.Config.Labels[ID] = c.ContainerCfg.Id

	return &opts, nil
}

func appendVolumesFrom(client *dockerClient.Client, containerCfg *config.ContainerConfig, opts *dockerClient.CreateContainerOptions) error {
	if !containerCfg.MigrateVolumes {
		return nil
	}

	container, err := getByLabel(client, ID, containerCfg.Id)
	if err != nil || container == nil {
		return err
	}

	if opts.HostConfig.VolumesFrom == nil {
		opts.HostConfig.VolumesFrom = []string{container.ID}
	} else {
		opts.HostConfig.VolumesFrom = append(opts.HostConfig.VolumesFrom, container.ID)
	}

	return nil
}

func (c *Container) start(create_only, wait bool) *Container {
	c.Lookup()
	c.Stage()

	if c.Err != nil {
		return c
	}

	client, err := NewClient(c.dockerHost)
	if err != nil {
		return c.returnErr(err)
	}

	container := c.Container
	created := false

	opts, err := c.getCreateOpts(client)
	if err != nil {
		return c.returnErr(err)
	}

	if c.Exists() && c.remove {
		log.Debugf("Deleting container %s", c.Container.ID)
		c.Delete()

		if c.Err != nil {
			return c
		}

		c.Reset().Lookup()
		if c.Err != nil {
			return c
		}
	}

	if !c.Exists() {
		err = c.renameOld(client, opts)
		if err != nil {
			return c.returnErr(err)
		}

		err := appendVolumesFrom(client, c.ContainerCfg, opts)
		if err != nil {
			return c.returnErr(err)
		}

		container, err = client.CreateContainer(*opts)
		created = true
		if err != nil {
			return c.returnErr(err)
		}
	}

	c.Container = container

	hostConfig := c.Container.HostConfig
	if created {
		hostConfig = opts.HostConfig
	}

	if create_only {
		return c
	}

	if !c.Container.State.Running {
		if !created {
			err = c.renameOld(client, opts)
			if err != nil {
				return c.returnErr(err)
			}
		}

		err = c.renameCurrent(client)
		if err != nil {
			return c.returnErr(err)
		}

		err = client.StartContainer(c.Container.ID, hostConfig)
		if err != nil {
			return c.returnErr(err)
		}
	}

	if !c.detach && wait {
		_, c.Err = client.WaitContainer(c.Container.ID)
		return c
	}

	return c
}
