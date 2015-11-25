package docker

import (
	"github.com/docker/libcompose/project"
	shlex "github.com/flynn/go-shlex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCommand(t *testing.T) {
	exp := []string{"sh", "-c", "exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}"}
	cmd, err := shlex.Split("sh -c 'exec /opt/bin/flanneld -logtostderr=true -iface=${NODE_IP}'")
	assert.Nil(t, err)
	assert.Equal(t, exp, cmd)
}

func TestParseBindsAndVolumes(t *testing.T) {
	cfg, hostCfg, err := Convert(&project.ServiceConfig{
		Volumes: []string{"/foo", "/home:/home", "/bar/baz", "/usr/lib:/usr/lib:ro"},
	})
	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{"/foo": {}, "/bar/baz": {}}, cfg.Volumes)
	assert.Equal(t, []string{"/home:/home", "/usr/lib:/usr/lib:ro"}, hostCfg.Binds)
}
