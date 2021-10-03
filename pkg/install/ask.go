package install

import (
	"os/exec"
	"strings"

	"github.com/rancher/os/pkg/config"
	"github.com/rancher/os/pkg/questions"
	"github.com/rancher/os/pkg/util"
)

func Ask(cfg *config.Config) error {
	if cfg.Rancher.Install.Silent {
		return nil
	}

	if err := AskInstallDevice(cfg); err != nil {
		return err
	}

	if err := AskConfigURL(cfg); err != nil {
		return err
	}

	if cfg.Rancher.Install.ConfigURL == "" {
		if err := AskGithub(cfg); err != nil {
			return err
		}

		if err := AskPassword(cfg); err != nil {
			return err
		}

		if err := AskServerAgent(cfg); err != nil {
			return err
		}
	}

	return nil
}

func AskInstallDevice(cfg *config.Config) error {
	if cfg.Rancher.Install.Device != "" {
		return nil
	}

	output, err := exec.Command("/bin/sh", "-c", "lsblk -r -o NAME,TYPE | grep -w disk | grep -v fd0 | awk '{print $1}'").CombinedOutput()
	if err != nil {
		return err
	}
	fields := strings.Fields(string(output))
	i, err := questions.PromptFormattedOptions("Installation target. Device will be formatted", -1, fields...)
	if err != nil {
		return err
	}

	cfg.Rancher.Install.Device = "/dev/" + fields[i]
	return nil
}

func AskToken(cfg *config.Config, server bool) error {
	var (
		token string
		err   error
	)

	if cfg.Rancher.Install.Token != "" {
		return nil
	}

	msg := "Token or cluster secret"
	if server {
		msg += " (optional)"
	}
	if server {
		token, err = questions.PromptOptional(msg+": ", "")
	} else {
		token, err = questions.Prompt(msg+": ", "")
	}
	cfg.Rancher.Install.Token = token

	return err
}

func isServer(cfg *config.Config) (bool, error) {
	opts := []string{"server", "agent"}
	i, err := questions.PromptFormattedOptions("Run as server or agent?", 0, opts...)
	if err != nil {
		return false, err
	}

	return i == 0, nil
}

func AskServerAgent(cfg *config.Config) error {
	if cfg.Rancher.Install.ServerURL != "" {
		return nil
	}

	server, err := isServer(cfg)
	if err != nil {
		return err
	}

	if server {
		return AskToken(cfg, true)
	}

	url, err := questions.Prompt("URL of server: ", "")
	if err != nil {
		return err
	}
	cfg.Rancher.Install.ServerURL = url

	return AskToken(cfg, false)
}

func AskPassword(cfg *config.Config) error {
	if cfg.Rancher.Install.Silent || cfg.Rancher.Install.Password != "" {
		return nil
	}

	var (
		ok   = false
		err  error
		pass string
	)

	for !ok {
		pass, ok, err = util.PromptPassword()
		if err != nil {
			return err
		}
	}

	if pass != "" {
		pass, err = util.GetEncryptedPasswd(pass)
		if err != nil {
			return err
		}
	}

	cfg.Rancher.Install.Password = pass
	return nil
}

func AskGithub(cfg *config.Config) error {
	if len(cfg.SSHAuthorizedKeys) > 0 || cfg.Rancher.Install.Password != "" {
		return nil
	}

	ok, err := questions.PromptBool("Authorize GitHub users to root SSH?", false)
	if !ok || err != nil {
		return err
	}

	str, err := questions.Prompt("Comma separated list of GitHub users to authorize: ", "")
	if err != nil {
		return err
	}

	for _, s := range strings.Split(str, ",") {
		cfg.SSHAuthorizedKeys = append(cfg.SSHAuthorizedKeys, "github:"+strings.TrimSpace(s))
	}

	return nil
}

func AskConfigURL(cfg *config.Config) error {
	if cfg.Rancher.Install.ConfigURL != "" {
		return nil
	}

	ok, err := questions.PromptBool("Configure system using an cloud-config file?", false)
	if err != nil {
		return err
	}

	if !ok {
		return nil
	}

	str, err := questions.Prompt("cloud-config file location (file path or http URL): ", "")
	if err != nil {
		return err
	}

	cfg.Rancher.Install.ConfigURL = str
	return nil
}
