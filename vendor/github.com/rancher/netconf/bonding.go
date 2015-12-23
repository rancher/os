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
	err  error
	name string
}

func (b *Bonding) Error() error {
	return b.err
}

func (b *Bonding) init() {
	_, err := os.Stat(bondingMasters)
	if os.IsNotExist(err) {
		logrus.Info("Loading bonding kernel module")
		cmd := exec.Command("modprobe", "bonding")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdin
		b.err = cmd.Run()
		if b.err != nil {
			for i := 0; i < 30; i++ {
				if _, err := os.Stat(bondingMasters); err == nil {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
	_, err = os.Stat(bondingMasters)
	if err != nil {
		b.err = err
	}
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

func (b *Bonding) AddSlave(slave string) {
	if b.err != nil {
		return
	}

	if ok, err := contains(base+b.name+"/bonding/slaves", slave); err != nil {
		b.err = err
		return
	} else if ok {
		return
	}

	link, err := netlink.LinkByName(slave)
	if err != nil {
		b.err = err
		return
	}

	b.err = netlink.LinkSetDown(link)
	if b.err != nil {
		return
	}

	p := base + b.name + "/bonding/slaves"
	logrus.Infof("Adding slave %s to master %s", slave, b.name)
	b.err = ioutil.WriteFile(p, []byte("+"+slave), 0644)
}

func (b *Bonding) Opt(key, value string) {
	if b.err != nil {
		return
	}

	p := base + b.name + "/bonding/" + key
	b.err = ioutil.WriteFile(p, []byte(value), 0644)
	if b.err != nil {
		logrus.Errorf("Failed to set %s=%s on %s", key, value, b.name)
	} else {
		logrus.Infof("Set %s=%s on %s", key, value, b.name)
	}
}

func (b *Bonding) Clear() {
	b.err = nil
}

func Bond(name string) *Bonding {
	b := &Bonding{name: name}
	b.init()
	if b.err != nil {
		return b
	}

	if ok, err := contains(bondingMasters, name); err != nil {
		b.err = err
		return b
	} else if ok {
		return b
	}

	logrus.Infof("Creating bond %s", name)
	b.err = ioutil.WriteFile(bondingMasters, []byte("+"+name), 0644)
	return b
}
