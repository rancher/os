package install

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rancher/os2/pkg/config"
	"github.com/rancher/os2/pkg/questions"
	"sigs.k8s.io/yaml"
)

func Run(automatic bool, configFile string, powerOff bool, silent bool) error {
	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}

	if powerOff {
		cfg.RancherOS.Install.PowerOff = true
	}

	if automatic && !cfg.RancherOS.Install.Automatic {
		return nil
	}

	if silent {
		cfg.RancherOS.Install.Automatic = true
	}

	err = Ask(&cfg)
	if err != nil {
		return err
	}

	tempFile, err := ioutil.TempFile("", "ros-install")
	if err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return runInstall(cfg, tempFile.Name())
}

func runInstall(cfg config.Config, output string) error {
	installBytes, err := config.PrintInstall(cfg)
	if err != nil {
		return err
	}

	if !cfg.RancherOS.Install.Automatic {
		val, err := questions.PromptBool("\nConfiguration\n"+"-------------\n\n"+
			string(installBytes)+
			"\nYour disk will be formatted and installed with the above configuration.\nContinue?", false)
		if err != nil || !val {
			return err
		}
	}

	if cfg.RancherOS.Install.ConfigURL == "" && !cfg.RancherOS.Install.Automatic {
		yip := config.YipConfig{
			Rancherd: config.Rancherd{
				Server: cfg.RancherOS.Install.ServerURL,
				Token:  cfg.RancherOS.Install.Token,
				Role:   cfg.RancherOS.Install.Role,
			},
		}
		if cfg.RancherOS.Install.Password != "" || len(cfg.SSHAuthorizedKeys) > 0 {
			yip.Stages = map[string][]config.Stage{
				"network": {{
					Users: map[string]config.User{
						"root": {
							Name:              "root",
							PasswordHash:      cfg.RancherOS.Install.Password,
							SSHAuthorizedKeys: cfg.SSHAuthorizedKeys,
						},
					}},
				}}
			cfg.RancherOS.Install.Password = ""
		}

		data, err := yaml.Marshal(yip)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(output+".yip", data, 0600); err != nil {
			return err
		}
		cfg.RancherOS.Install.ConfigURL = output + ".yip"
	}

	if err := config.ToFile(cfg, output); err != nil {
		return err
	}
	cfg.RancherOS.Install.ConfigURL = output

	ev, err := config.ToEnv(cfg)
	if err != nil {
		return err
	}

	printEnv(cfg)

	cmd := exec.Command("cos-installer")
	cmd.Env = append(os.Environ(), ev...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printEnv(cfg config.Config) {
	if cfg.RancherOS.Install.Password != "" {
		cfg.RancherOS.Install.Password = "<removed>"
	}

	ev2, err := config.ToEnv(cfg)
	if err != nil {
		return
	}

	fmt.Println("Install environment:")
	for _, ev := range ev2 {
		fmt.Println(ev)
	}
}
