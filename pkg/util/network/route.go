package network

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rancher/os/pkg/log"
	"github.com/rancher/os/pkg/util"
)

const (
	DefaultRoutesCheckTimeout = 20000 // 20 second

	ipv4RouteFile = "/proc/net/route"
	ipv6RouteFile = "/proc/net/ipv6_route"

	ipv4DefaultGWFlags = "0003"
	ipv6DefaultGWFlags = "00450003"
)

func checkIPv4GW() bool {
	if _, err := os.Stat(ipv4RouteFile); os.IsNotExist(err) {
		return false
	}
	file, err := os.Open(ipv4RouteFile)
	if err != nil {
		return false
	}
	defer file.Close()
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')
		if err == io.EOF {
			break
		}
		//ignore the headers in the route info
		if strings.HasPrefix(line, "Iface") {
			continue
		}
		fields := strings.Fields(line)
		// Interested in fields:
		//  3 - flags
		if fields[3] == ipv4DefaultGWFlags {
			return true
		}
	}

	return false
}

func checkIPv6GW() bool {
	if _, err := os.Stat(ipv6RouteFile); os.IsNotExist(err) {
		return false
	}
	file, err := os.Open(ipv6RouteFile)
	if err != nil {
		return false
	}
	defer file.Close()
	scanner := bufio.NewReader(file)
	for {
		line, err := scanner.ReadString('\n')
		if err == io.EOF {
			break
		}
		fields := strings.Fields(line)
		// Interested in fields:
		//  3 - flags
		if fields[3] == ipv6DefaultGWFlags {
			return true
		}
	}

	return false
}

func checkAllDefaultGW() bool {
	return checkIPv4GW() || checkIPv6GW()
}

func AllDefaultGWOK(timeout int) error {
	backoff := util.Backoff{
		MaxMillis: timeout,
	}
	defer backoff.Close()

	var err error
	for ok := range backoff.Start() {
		if !ok {
			err = fmt.Errorf("Timeout waiting for the default gateway ready")
			break
		}
		if checkAllDefaultGW() {
			break
		}
		log.Info("Waiting for the default gateway ready")
	}

	if err != nil {
		return err
	}

	log.Info("The default gateway is ready")

	return nil
}
