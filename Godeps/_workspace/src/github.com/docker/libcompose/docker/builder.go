package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/utils"
	"github.com/docker/libcompose/project"
	"github.com/samalba/dockerclient"
)

type Builder interface {
	Build(p *project.Project, service project.Service) (string, error)
}

type DaemonBuilder struct {
	context *Context
}

func NewDaemonBuilder(context *Context) *DaemonBuilder {
	return &DaemonBuilder{
		context: context,
	}
}

func (d *DaemonBuilder) Build(p *project.Project, service project.Service) (string, error) {
	if service.Config().Build == "" {
		return service.Config().Image, nil
	}

	tag := fmt.Sprintf("%s_%s", p.Name, service.Name())
	context, err := CreateTar(p, service.Name())
	if err != nil {
		return "", err
	}

	defer context.Close()

	client := d.context.ClientFactory.Create(service)

	logrus.Infof("Building %s...", tag)
	output, err := client.BuildImage(&dockerclient.BuildImage{
		Context:  context,
		RepoName: tag,
		Remove:   true,
	})
	if err != nil {
		return "", err
	}

	defer output.Close()

	// Don't really care about errors in the scanner
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		text := scanner.Text()
		data := map[string]interface{}{}
		err := json.Unmarshal([]byte(text), &data)
		if stream, ok := data["stream"]; ok && err == nil {
			fmt.Print(stream)
		}
	}

	return tag, nil
}

func CreateTar(p *project.Project, name string) (io.ReadCloser, error) {
	// This code was ripped off from docker/api/client/build.go

	serviceConfig := p.Configs[name]
	root := serviceConfig.Build
	dockerfileName := filepath.Join(root, serviceConfig.Dockerfile)

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	filename := dockerfileName

	if dockerfileName == "" {
		// No -f/--file was specified so use the default
		dockerfileName = api.DefaultDockerfileName
		filename = filepath.Join(absRoot, dockerfileName)

		// Just to be nice ;-) look for 'dockerfile' too but only
		// use it if we found it, otherwise ignore this check
		if _, err = os.Lstat(filename); os.IsNotExist(err) {
			tmpFN := path.Join(absRoot, strings.ToLower(dockerfileName))
			if _, err = os.Lstat(tmpFN); err == nil {
				dockerfileName = strings.ToLower(dockerfileName)
				filename = tmpFN
			}
		}
	}

	origDockerfile := dockerfileName // used for error msg
	if filename, err = filepath.Abs(filename); err != nil {
		return nil, err
	}

	// Now reset the dockerfileName to be relative to the build context
	dockerfileName, err = filepath.Rel(absRoot, filename)
	if err != nil {
		return nil, err
	}

	// And canonicalize dockerfile name to a platform-independent one
	dockerfileName, err = archive.CanonicalTarNameForPath(dockerfileName)
	if err != nil {
		return nil, fmt.Errorf("Cannot canonicalize dockerfile path %s: %v", dockerfileName, err)
	}

	if _, err = os.Lstat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("Cannot locate Dockerfile: %s", origDockerfile)
	}
	var includes = []string{"."}

	excludes, err := utils.ReadDockerIgnore(path.Join(root, ".dockerignore"))
	if err != nil {
		return nil, err
	}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	keepThem1, _ := fileutils.Matches(".dockerignore", excludes)
	keepThem2, _ := fileutils.Matches(dockerfileName, excludes)
	if keepThem1 || keepThem2 {
		includes = append(includes, ".dockerignore", dockerfileName)
	}

	if err := utils.ValidateContextDirectory(root, excludes); err != nil {
		return nil, fmt.Errorf("Error checking context is accessible: '%s'. Please check permissions and try again.", err)
	}

	options := &archive.TarOptions{
		Compression:     archive.Uncompressed,
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
	}

	return archive.TarWithOptions(root, options)
}
