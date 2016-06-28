package daemon

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/container"
	"github.com/opencontainers/specs/specs-go"
	"github.com/pkg/errors"
)

var (
	prestartDir  = "/etc/docker/hooks/prestart.d"
	poststartDir = "/etc/docker/hooks/poststart.d"
	poststopDir  = "/etc/docker/hooks/poststop.d"
)

func loadHooks(hookDir string) ([]specs.Hook, error) {
	files, err := ioutil.ReadDir(hookDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "read hooks dir failed "+hookDir)
	}

	result := []specs.Hook{}

	for _, f := range files {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}

		of, err := os.Open(path.Join(hookDir, f.Name()))
		if err != nil {
			return nil, errors.Wrap(err, "failed to open "+f.Name())
		}
		defer of.Close()

		var spec specs.Hook
		if err := json.NewDecoder(of).Decode(&spec); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshall "+f.Name())
		}

		result = append(result, spec)
	}

	return result, nil
}

func addHooks(c *container.Container, spec *specs.Spec) error {
	prestart, err := loadHooks(prestartDir)
	if err != nil {
		return err
	}

	poststart, err := loadHooks(poststartDir)
	if err != nil {
		return err
	}

	poststop, err := loadHooks(poststopDir)
	if err != nil {
		return err
	}

	configPath, err := c.ConfigPath()
	if err != nil {
		return errors.Wrap(err, "config path")
	}

	hostConfigPath, err := c.HostConfigPath()
	if err != nil {
		return errors.Wrap(err, "host config path")
	}

	spec.Hooks.Prestart = appendHooksForContainer(configPath, hostConfigPath, c, prestart)
	spec.Hooks.Poststart = appendHooksForContainer(configPath, hostConfigPath, c, poststart)
	spec.Hooks.Poststop = appendHooksForContainer(configPath, hostConfigPath, c, poststop)

	return nil
}

func appendHooksForContainer(configPath, hostConfigPath string, c *container.Container, hooks []specs.Hook) []specs.Hook {
	result := []specs.Hook{}
	for _, hook := range hooks {
		hook.Env = append(hook.Env,
			fmt.Sprintf("DOCKER_CONFIG=%s", configPath),
			fmt.Sprintf("DOCKER_HOST_CONFIG=%s", hostConfigPath))

		result = append(result, hook)
	}
	return result
}
