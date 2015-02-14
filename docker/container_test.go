package docker

import (
	"testing"

	"github.com/rancherio/os/config"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	assert := require.New(t)

	hash, err := getHash(&config.ContainerConfig{
		Id:  "id",
		Cmd: []string{"1", "2", "3"},
	})
	assert.NoError(err, "")

	hash2, err := getHash(&config.ContainerConfig{
		Id:  "id2",
		Cmd: []string{"1", "2", "3"},
	})
	assert.NoError(err, "")

	hash3, err := getHash(&config.ContainerConfig{
		Id:  "id3",
		Cmd: []string{"1", "2", "3", "4"},
	})
	assert.NoError(err, "")

	assert.Equal("44096e94ed438ccda24e459412147441a376ea1c", hash, "")
	assert.NotEqual(hash, hash2, "")
	assert.NotEqual(hash2, hash3, "")
	assert.NotEqual(hash, hash3, "")
}

func TestParse(t *testing.T) {
	assert := require.New(t)

	cfg := &config.ContainerConfig{
		Cmd: []string{
			"--name", "c1",
			"-d",
			"--rm",
			"--privileged",
			"test/image",
			"arg1",
			"arg2",
		},
	}

	c := NewContainer(nil, cfg).Parse()

	assert.NoError(c.Err, "")
	assert.Equal(cfg.Id, "c1", "Id doesn't match")
	assert.Equal(c.Name, "c1", "Name doesn't match")
	assert.True(c.remove, "Remove doesn't match")
	assert.True(c.detach, "Detach doesn't match")
	assert.Equal(len(c.Config.Cmd), 2, "Args doesn't match")
	assert.Equal(c.Config.Cmd[0], "arg1", "Arg1 doesn't match")
	assert.Equal(c.Config.Cmd[1], "arg2", "Arg2 doesn't match")
	assert.True(c.HostConfig.Privileged, "Privileged doesn't match")
}

func TestStart(t *testing.T) {
	assert := require.New(t)

	c := NewContainer(nil, &config.ContainerConfig{
		Cmd: []string{"--pid=host", "--privileged", "--rm", "busybox", "echo", "hi"},
	}).Parse().Start().Lookup()

	assert.NoError(c.Err, "")

	assert.True(c.HostConfig.Privileged, "")
	assert.True(c.container.HostConfig.Privileged, "")
	assert.Equal("host", c.container.HostConfig.PidMode, "")

	c.Delete()
}

func TestLookup(t *testing.T) {
	assert := require.New(t)

	cfg := &config.ContainerConfig{
		Cmd: []string{"--rm", "busybox", "echo", "hi"},
	}
	c := NewContainer(nil, cfg).Parse().Start()

	cfg2 := &config.ContainerConfig{
		Cmd: []string{"--rm", "busybox", "echo", "hi2"},
	}
	c2 := NewContainer(nil, cfg2).Parse().Start()

	assert.NoError(c.Err, "")
	assert.NoError(c2.Err, "")

	c1Lookup := NewContainer(nil, cfg).Lookup()
	c2Lookup := NewContainer(nil, cfg2).Lookup()

	assert.NoError(c1Lookup.Err, "")
	assert.NoError(c2Lookup.Err, "")

	assert.Equal(c.container.ID, c1Lookup.container.ID, "")
	assert.Equal(c2.container.ID, c2Lookup.container.ID, "")

	c.Delete()
	c2.Delete()
}

func TestDelete(t *testing.T) {
	assert := require.New(t)

	c := NewContainer(nil, &config.ContainerConfig{
		Cmd: []string{"--rm", "busybox", "echo", "hi"},
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
