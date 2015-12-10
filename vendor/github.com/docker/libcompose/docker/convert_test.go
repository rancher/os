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
	bashCmd := "bash"
	fooLabel := "foo.label"
	fooLabelValue := "service.config.value"
	sc := &project.ServiceConfig{
		Entrypoint: project.NewCommand(bashCmd),
		Volumes:    []string{"/foo", "/home:/home", "/bar/baz", "/usr/lib:/usr/lib:ro"},
		Labels:     project.NewSliceorMap(map[string]string{fooLabel: "service.config.value"}),
	}
	cfg, hostCfg, err := Convert(sc)
	assert.Nil(t, err)
	assert.Equal(t, map[string]struct{}{"/foo": {}, "/bar/baz": {}}, cfg.Volumes)
	assert.Equal(t, []string{"/home:/home", "/usr/lib:/usr/lib:ro"}, hostCfg.Binds)

	cfg.Labels[fooLabel] = "FUN"
	cfg.Entrypoint[0] = "less"

	assert.Equal(t, fooLabelValue, sc.Labels.MapParts()[fooLabel])
	assert.Equal(t, "FUN", cfg.Labels[fooLabel])

	assert.Equal(t, []string{bashCmd}, sc.Entrypoint.Slice())
	assert.Equal(t, []string{"less"}, cfg.Entrypoint)
}
