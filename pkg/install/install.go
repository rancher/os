package install

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rancher/os/pkg/config"
	"github.com/rancher/os/pkg/questions"
	"sigs.k8s.io/yaml"
)

func Run(automatic bool) error {
	cfg, err := config.ReadConfig()
	if err != nil {
		return err
	}

	if automatic && !cfg.Elemental.Install.Automatic {
		return nil
	} else if automatic {
		cfg.Elemental.Install.Silent = true
	}

	err = Ask(&cfg)
	if err != nil {
		return err
	}

	tempFile, err := ioutil.TempFile("", "elemental-install")
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

	if !cfg.Elemental.Install.Silent {
		val, err := questions.PromptBool("\nConfiguration\n"+"-------------\n\n"+
			string(installBytes)+
			"\nYour disk will be formatted and installed with the above configuration.\nContinue?", false)
		if err != nil || !val {
			return err
		}
	}

	if cfg.Elemental.Install.ConfigURL == "" {
		yip := config.YipConfig{
			Rancherd: config.Rancherd{
				Server: cfg.Elemental.Install.ServerURL,
				Token:  cfg.Elemental.Install.Token,
			},
		}
		if cfg.Elemental.Install.ServerURL == "" {
			yip.Rancherd.Role = "cluster-init"
		} else {
			yip.Rancherd.Role = "agent"
		}
		if cfg.Elemental.Install.Password != "" || len(cfg.SSHAuthorizedKeys) > 0 {
			yip.Stages = map[string][]config.Stage{
				"initramfs": {{
					Users: map[string]config.User{
						"root": {
							Name:              "root",
							PasswordHash:      cfg.Elemental.Install.Password,
							SSHAuthorizedKeys: cfg.SSHAuthorizedKeys,
						},
					}},
				}}
			cfg.Elemental.Install.Password = ""
		}

		data, err := yaml.Marshal(yip)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(output+".yip", data, 0600); err != nil {
			return err
		}
		cfg.Elemental.Install.ConfigURL = output + ".yip"
	}

	ev, err := config.ToEnv(cfg)
	if err != nil {
		return err
	}

	cmd := exec.Command("cos-installer")
	cmd.Env = append(os.Environ(), ev...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
