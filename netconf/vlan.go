package netconf

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
)

type VlanDefinition struct {
	Id   int
	Name string
}

type Vlan struct {
	name string
	link netlink.Link
	id   int
}

func NewVlan(link netlink.Link, name string, id int) (*Vlan, error) {
	if name == "" {
		name = fmt.Sprintf("%s.%d", link.Attrs().Name, id)
	}

	v := &Vlan{
		name: name,
		link: link,
		id:   id,
	}
	return v, v.init()
}

func (v *Vlan) init() error {
	link, err := netlink.LinkByName(v.name)
	if err == nil {
		if _, ok := link.(*netlink.Vlan); !ok {
			return fmt.Errorf("%s is not a VLAN device", v.name)
		}
		return nil
	}

	vlan := netlink.Vlan{}
	vlan.ParentIndex = v.link.Attrs().Index
	vlan.Name = v.name
	vlan.VlanId = v.id

	return netlink.LinkAdd(&vlan)
}

func ParseVlanDefinitions(vlans string) ([]VlanDefinition, error) {
	vlans = strings.TrimSpace(vlans)
	if vlans == "" {
		return nil, nil
	}

	result := []VlanDefinition{}

	for _, vlan := range strings.Split(vlans, ",") {
		idName := strings.SplitN(strings.TrimSpace(vlan), ":", 2)
		id, err := strconv.Atoi(idName[0])
		if err != nil {
			return nil, fmt.Errorf("Invalid format in %s: %v", vlans, err)
		}

		def := VlanDefinition{
			Id: id,
		}

		if len(idName) > 1 {
			def.Name = idName[1]
		}

		result = append(result, def)
	}

	return result, nil
}
