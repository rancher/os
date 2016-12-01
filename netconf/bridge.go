package netconf

import (
	"fmt"

	"github.com/vishvananda/netlink"
)

type Bridge struct {
	name string
}

func NewBridge(name string) (*Bridge, error) {
	b := &Bridge{name: name}
	return b, b.init()
}

func (b *Bridge) init() error {
	link, err := netlink.LinkByName(b.name)
	if err == nil {
		if _, ok := link.(*netlink.Bridge); !ok {
			return fmt.Errorf("%s is not a bridge device", b.name)
		}
		return nil
	}

	bridge := netlink.Bridge{}
	bridge.LinkAttrs.Name = b.name

	return netlink.LinkAdd(&bridge)
}

func (b *Bridge) AddLink(link netlink.Link) error {
	existing, err := netlink.LinkByName(b.name)
	if err != nil {
		return err
	}

	if bridge, ok := existing.(*netlink.Bridge); ok {
		if link.Attrs().MasterIndex != bridge.Index {
			return netlink.LinkSetMaster(link, bridge)
		}
	} else {
		return fmt.Errorf("%s is not a bridge", b.name)
	}

	return nil
}
