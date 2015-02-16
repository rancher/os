package sysinit

import (
	"os"
	"os/exec"
	"path"
	"syscall"

	log "github.com/Sirupsen/logrus"
	dockerClient "github.com/fsouza/go-dockerclient"
	"github.com/rancherio/os/config"
	"github.com/rancherio/os/docker"
	initPkg "github.com/rancherio/os/init"
	"github.com/rancherio/os/util"
)

func SysInit() {
	if err := sysInit(); err != nil {
		log.Fatal(err)
	}
}

func importImage(client *dockerClient.Client, name, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	log.Debugf("Importing image for %s", fileName)
	repo, tag := dockerClient.ParseRepositoryTag(name)
	return client.ImportImage(dockerClient.ImportImageOptions{
		Source:      "-",
		Repository:  repo,
		Tag:         tag,
		InputStream: file,
	})
}

func hasImage(name string) bool {
	stamp := path.Join(initPkg.STATE, name)
	if _, err := os.Stat(stamp); os.IsNotExist(err) {
		return false
	}
	return true
}

func findImages(cfg *config.Config) ([]string, error) {
	log.Debugf("Looking for images at %s", cfg.ImagesPath)

	result := []string{}

	dir, err := os.Open(cfg.ImagesPath)
	if os.IsNotExist(err) {
		log.Debugf("Not loading images, %s does not exist")
		return result, nil
	}
	if err != nil {
		return nil, err
	}

	defer dir.Close()

	files, err := dir.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	for _, fileName := range files {
		if ok, _ := path.Match(cfg.ImagesPattern, fileName); ok {
			log.Debugf("Found %s", fileName)
			result = append(result, fileName)
		}
	}

	return result, nil
}

func loadImages(cfg *config.Config) error {
	images, err := findImages(cfg)
	if err != nil || len(images) == 0 {
		return err
	}

	client, err := docker.NewClient(cfg)
	if err != nil {
		return err
	}

	for _, image := range images {
		if hasImage(image) {
			continue
		}

		inputFileName := path.Join(cfg.ImagesPath, image)
		input, err := os.Open(inputFileName)
		if err != nil {
			return err
		}

		defer input.Close()

		log.Debugf("Loading images from %s", inputFileName)
		err = client.LoadImage(dockerClient.LoadImageOptions{
			InputStream: input,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func runContainers(cfg *config.Config) error {
	containerConfigs := cfg.SystemContainers
	if cfg.Rescue {
		log.Debug("Running rescue container")
		containerConfigs = []config.ContainerConfig{cfg.RescueContainer}
	}

	for _, containerConfig := range containerConfigs {
		container := docker.NewContainer(cfg, &containerConfig)
		container.Parse()

		if util.Contains(cfg.Disable, containerConfig.Id) {
			log.Debugf("%s is disabled : %v", containerConfig.Id, cfg.Disable)
			continue
		}

		if containerConfig.Id == cfg.ConsoleContainer {
			if util.IsRunningInTty() {
				container.Config.Tty = true
				container.Config.AttachStdin = true
				container.Config.AttachStdout = true
				container.Config.AttachStderr = true
			}
		}

		container.StartAndWait()
		log.Debugf("Running %s", containerConfig.Id)

		if container.Err != nil {
			log.Errorf("Failed to run %v: %v", containerConfig.Id, container.Err)
		}
	}

	return nil
}

func launchConsole(cfg *config.Config) error {
	if !util.IsRunningInTty() {
		return nil
	}

	log.Debugf("Attaching to console")
	cmd := exec.Command("docker", "attach", "console")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Start()

	return cmd.Wait()
	//console := cfg.GetContainerById(cfg.ConsoleContainer)
	//if console == nil {
	//	return nil
	//}

	//c, err := docker.ParseContainer(cfg, console)
	//if err != nil {
	//	return err
	//}

	//return c.Attach()

	//return nil
}

func sysInit() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	initFuncs := []config.InitFunc{
		loadImages,
		runContainers,
		func(cfg *config.Config) error {
			syscall.Sync()
			return nil
		},
		//launchConsole,
	}

	return config.RunInitFuncs(cfg, initFuncs)
}
