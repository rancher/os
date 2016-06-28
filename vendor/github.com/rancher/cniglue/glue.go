package glue

import (
	"os"

	"github.com/Sirupsen/logrus"
)

func Main() {
	var err error
	if len(os.Args) < 2 || os.Args[1] == "prestart" {
		err = Prestart()
	} else if os.Args[1] == "poststop" {
		err = Poststop()
	}
	if err != nil {
		logrus.Fatal(err)
	}
}

func Prestart() error {
	state, err := ReadState()
	if err != nil {
		return err
	}

	if err := SetupResolvConf(state); err != nil {
		return err
	}

	cniResult, err := CNIAdd(state)
	if err != nil {
		return err
	}

	return SetupHosts(state, cniResult)
}

func Poststop() error {
	state, err := ReadState()
	if err != nil {
		return err
	}

	return CNIDel(state)
}
