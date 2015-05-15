package docker

import (
	"strings"
	"testing"

	"github.com/rancherio/os/config"
	"github.com/stretchr/testify/require"

	dockerClient "github.com/fsouza/go-dockerclient"
)

func TestHash(t *testing.T) {
	assert := require.New(t)

	hash, err := getHash(&config.ContainerConfig{
		Id:  "id",
		Cmd: "1 2 3",
	})
	assert.NoError(err, "")

	hash2, err := getHash(&config.ContainerConfig{
		Id:  "id2",
		Cmd: "1 2 3",
	})
	assert.NoError(err, "")

	hash3, err := getHash(&config.ContainerConfig{
		Id:  "id3",
		Cmd: "1 2 3 4",
	})
	assert.NoError(err, "")

	assert.Equal("510b68938cba936876588b0143093a5850d4a142", hash, "")
	assert.NotEqual(hash, hash2, "")
	assert.NotEqual(hash2, hash3, "")
	assert.NotEqual(hash, hash3, "")
}

func TestParse(t *testing.T) {
	assert := require.New(t)

	cfg := &config.ContainerConfig{
		Cmd: "--name c1 " +
			"-d " +
			"--rm " +
			"--privileged " +
			"test/image " +
			"arg1 " +
			"arg2 ",
	}

	c := NewContainer("", cfg).Parse()

	assert.NoError(c.Err, "")
	assert.Equal(cfg.Id, "c1", "Id doesn't match")
	assert.Equal(c.Name, "c1", "Name doesn't match")
	assert.True(c.remove, "Remove doesn't match")
	assert.True(c.detach, "Detach doesn't match")
	assert.Equal(c.Config.Cmd.Len(), 2, "Args doesn't match")
	assert.Equal(c.Config.Cmd.Slice()[0], "arg1", "Arg1 doesn't match")
	assert.Equal(c.Config.Cmd.Slice()[1], "arg2", "Arg2 doesn't match")
	assert.True(c.HostConfig.Privileged, "Privileged doesn't match")
}

func TestIdFromName(t *testing.T) {
	assert := require.New(t)

	cfg := &config.ContainerConfig{
		Cmd: "--name foo -v /test busybox echo hi",
	}

	assert.Equal("", cfg.Id)
	NewContainer(config.DOCKER_HOST, cfg)
	assert.Equal("foo", cfg.Id)
}

func TestMigrateVolumes(t *testing.T) {
	assert := require.New(t)

	c := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		Cmd: "--name foo -v /test busybox echo hi",
	}).Parse().Start().Lookup()

	assert.NoError(c.Err, "")

	test_path, ok := c.Container.Volumes["/test"]
	assert.True(ok, "")

	c2 := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		MigrateVolumes: true,
		Cmd:            "--name foo -v /test2 busybox echo hi",
	}).Parse().Start().Lookup()

	assert.NoError(c2.Err, "")

	assert.True(c2.Container != nil)

	_, ok = c2.Container.Volumes["/test2"]
	assert.True(ok, "")
	assert.Equal(test_path, c2.Container.Volumes["/test"])

	c.Delete()
	c2.Delete()
}

func TestRollback(t *testing.T) {
	assert := require.New(t)

	c := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		Cmd: "--name rollback busybox echo hi",
	}).Parse().Start().Lookup()

	assert.NoError(c.Err, "")
	assert.Equal("rollback", c.Container.Name)

	c2 := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		Cmd: "--name rollback busybox echo bye",
	}).Parse().Start().Lookup()

	assert.Equal("rollback", c2.Container.Name)
	assert.NoError(c2.Err, "")
	assert.NotEqual(c.Container.ID, c2.Container.ID)

	c3 := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		Cmd: "--name rollback busybox echo hi",
	}).Parse().Start().Lookup()

	assert.NoError(c3.Err, "")
	assert.Equal(c.Container.ID, c3.Container.ID)
	assert.Equal("rollback", c3.Container.Name)

	c2.Reset().Lookup()
	assert.NoError(c2.Err, "")
	assert.True(strings.HasPrefix(c2.Container.Name, "rollback-"))

	c.Delete()
	c2.Delete()
}

func TestStart(t *testing.T) {
	assert := require.New(t)

	c := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		Cmd: "--pid=host --privileged --rm busybox echo hi",
	}).Parse().Start().Lookup()

	assert.NoError(c.Err, "")

	assert.True(c.HostConfig.Privileged, "")
	assert.True(c.Container.HostConfig.Privileged, "")
	assert.Equal("host", c.Container.HostConfig.PidMode, "")

	c.Delete()
}

func TestLookup(t *testing.T) {
	assert := require.New(t)

	cfg := &config.ContainerConfig{
		Cmd: "--rm busybox echo hi",
	}
	c := NewContainer(config.DOCKER_HOST, cfg).Parse().Start()

	cfg2 := &config.ContainerConfig{
		Cmd: "--rm busybox echo hi2",
	}
	c2 := NewContainer(config.DOCKER_HOST, cfg2).Parse().Start()

	assert.NoError(c.Err, "")
	assert.NoError(c2.Err, "")

	c1Lookup := NewContainer(config.DOCKER_HOST, cfg).Lookup()
	c2Lookup := NewContainer(config.DOCKER_HOST, cfg2).Lookup()

	assert.NoError(c1Lookup.Err, "")
	assert.NoError(c2Lookup.Err, "")

	assert.Equal(c.Container.ID, c1Lookup.Container.ID, "")
	assert.Equal(c2.Container.ID, c2Lookup.Container.ID, "")

	c.Delete()
	c2.Delete()
}

func TestDelete(t *testing.T) {
	assert := require.New(t)

	c := NewContainer(config.DOCKER_HOST, &config.ContainerConfig{
		Cmd: "--rm busybox echo hi",
	}).Parse()

	assert.False(c.Exists())
	assert.NoError(c.Err, "")

	c.Start()
	assert.NoError(c.Err, "")
	c.Reset()
	assert.NoError(c.Err, "")

	assert.True(c.Exists())
	assert.NoError(c.Err, "")

	c.Delete()
	assert.NoError(c.Err, "")

	c.Reset()
	assert.False(c.Exists())
	assert.NoError(c.Err, "")
}

func TestDockerClientNames(t *testing.T) {
	assert := require.New(t)
	client, err := dockerClient.NewClient(config.DOCKER_HOST)

	assert.NoError(err, "")

	c, err := client.CreateContainer(dockerClient.CreateContainerOptions{
		Name: "foo",
		Config: &dockerClient.Config{
			Image: "ubuntu",
		},
	})

	assert.NoError(err, "")
	assert.Equal("foo", c.Name)

	c2, err := client.InspectContainer(c.ID)

	assert.NoError(err, "")
	assert.Equal("/foo", c2.Name)

	c2, err = inspect(client, c.ID)

	assert.NoError(err, "")
	assert.Equal("foo", c2.Name)

	client.RemoveContainer(dockerClient.RemoveContainerOptions{
		ID:    c2.ID,
		Force: true,
	})
}
