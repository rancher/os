package waitfornetwork

import (
	"github.com/rancher/os/cmd/network"
	"os"
	"os/signal"
	"syscall"
)

func handleTerm() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	<-c
	os.Exit(0)
}

func Main() {
	go handleTerm()
	if _, err := os.Stat(network.NETWORK_DONE); err == nil {
		os.Exit(0)
	}
	select {}
}
