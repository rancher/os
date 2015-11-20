package netconf

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"

	log "github.com/Sirupsen/logrus"

	"github.com/j-keck/arping"
	"github.com/vishvananda/netlink"
)

func AssignLinkLocalIP(link netlink.Link) error {
	ifaceName := link.Attrs().Name
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Error("could not get information about interface")
		return err
	}
	addrs, err := iface.Addrs()
	if err != nil {
		log.Error("Error fetching existing ip on interface")
	}
	for _, addr := range addrs {
		if addr.String()[:7] == "169.254" {
			log.Info("Link Local IP already set on interface")
			return nil
		}
	}
	randSource, err := getPseudoRandomGenerator(link.Attrs().HardwareAddr)
	if err != nil {
		return err
	}
	// try a random address upto 10 times
	for i := 0; i < 10; i++ {
		randGenerator := rand.New(*randSource)
		randomNum := randGenerator.Uint32()
		dstIP := getNewIPV4LLAddr(randomNum)
		if dstIP[2] == 0 || dstIP[2] == 255 {
			i--
			continue
		}
		_, _, err := arping.PingOverIfaceByName(dstIP, ifaceName)
		if err != nil {
			// this ip is not being used
			addr, err := netlink.ParseAddr(dstIP.String() + "/16")
			if err != nil {
				log.Errorf("error while parsing ipv4ll addr, err = %v", err)
				return err
			}
			if err := netlink.AddrAdd(link, addr); err != nil {
				log.Error("ipv4ll addr add failed")
				return err
			}
			log.Infof("Set %s on %s", dstIP.String(), link.Attrs().Name)
			return nil
		}
	}
	log.Error("Could not find a suitable ipv4ll")
	return fmt.Errorf("Could not find a suitable ipv4ll")
}

func getNewIPV4LLAddr(randomNum uint32) net.IP {
	byte1 := randomNum & 255 // use least significant 8 bits
	byte2 := randomNum >> 24 // use most significant 8 bits
	return []byte{169, 254, byte(byte1), byte(byte2)}
}

func getPseudoRandomGenerator(haAddr []byte) (*rand.Source, error) {
	seed, _ := binary.Varint(haAddr)
	src := rand.NewSource(seed)
	return &src, nil
}
