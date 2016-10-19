package netconf

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

const (
	base           = "/sys/class/net/"
	bondingMasters = "/sys/class/net/bonding_masters"
)

type Bonding struct {
	name string
}

func (b *Bonding) init() error {
	_, err := os.Stat(bondingMasters)
	if os.IsNotExist(err) {
		logrus.Info("Loading bonding kernel module")
		cmd := exec.Command("modprobe", "bonding")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdin
		err = cmd.Run()
		if err != nil {
			for i := 0; i < 30; i++ {
				if _, err = os.Stat(bondingMasters); err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	_, err = os.Stat(bondingMasters)
	return err
}

func contains(file, word string) (bool, error) {
	words, err := ioutil.ReadFile(file)
	if err != nil {
		return false, err
	}

	for _, s := range strings.Split(strings.TrimSpace(string(words)), " ") {
		if s == strings.TrimSpace(word) {
			return true, nil
		}
	}

	return false, nil
}

func (b *Bonding) linkDown() error {
	link, err := netlink.LinkByName(b.name)
	if err != nil {
		return err
	}

	return netlink.LinkSetDown(link)
}

func (b *Bonding) ListSlaves() ([]string, error) {
	file := base + b.name + "/bonding/slaves"
	words, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, s := range strings.Split(strings.TrimSpace(string(words)), " ") {
		if s != "" {
			result = append(result, s)
		}
	}
	return result, nil
}

func (b *Bonding) RemoveSlave(slave string) error {
	if ok, err := contains(base+b.name+"/bonding/slaves", slave); err != nil {
		return err
	} else if !ok {
		return nil
	}

	p := base + b.name + "/bonding/slaves"
	logrus.Infof("Removing slave %s from master %s", slave, b.name)
	return ioutil.WriteFile(p, []byte("-"+slave), 0644)
}

func (b *Bonding) AddSlave(slave string) error {
	if ok, err := contains(base+b.name+"/bonding/slaves", slave); err != nil {
		return err
	} else if ok {
		return nil
	}

	p := base + b.name + "/bonding/slaves"
	logrus.Infof("Adding slave %s to master %s", slave, b.name)
	return ioutil.WriteFile(p, []byte("+"+slave), 0644)
}

func (b *Bonding) Opt(key, value string) error {
	if key == "mode" {
		// Don't care about errors here
		b.linkDown()
		slaves, _ := b.ListSlaves()
		for _, slave := range slaves {
			b.RemoveSlave(slave)
		}
	}

	p := base + b.name + "/bonding/" + key
	if err := ioutil.WriteFile(p, []byte(value), 0644); err != nil {
		logrus.Errorf("Failed to set %s=%s on %s: %v", key, value, b.name, err)
		return err
	} else {
		logrus.Infof("Set %s=%s on %s", key, value, b.name)
	}

	return nil
}

func Bond(name string) (*Bonding, error) {
	b := &Bonding{name: name}
	if err := b.init(); err != nil {
		return nil, err
	}

	if ok, err := contains(bondingMasters, name); err != nil {
		return nil, err
	} else if ok {
		return b, nil
	}

	logrus.Infof("Creating bond %s", name)
	return b, ioutil.WriteFile(bondingMasters, []byte("+"+name), 0644)
}
