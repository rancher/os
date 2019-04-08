package netconf

import (
	"net"
	"testing"

	"github.com/vishvananda/netlink"
)

type mockLink struct {
	attrs netlink.LinkAttrs
	t     string
}

func (l mockLink) Attrs() *netlink.LinkAttrs {
	return &l.attrs
}

func (l mockLink) Type() string {
	return l.t
}

func TestFindMatch(t *testing.T) {
	testCases := []struct {
		match    string
		mac      string
		t        string
		name     string
		bond     string
		expected bool
	}{
		{
			"mac:aa:bb:cc:dd:ee:ff",
			"aa:bb:cc:dd:ee:ff",
			"fake",
			"eth0",
			"bond0",
			true,
		},
		{
			"mac:aa:bb:cc:*",
			"aa:bb:cc:12:34:56",
			"fake",
			"eth0",
			"bond0",
			true,
		},
		{
			"mac:aa:bb:cc:*",
			"11:bb:cc:dd:ee:ff",
			"fake",
			"eth0",
			"bond0",
			false,
		},
		{
			"mac:aa:bb:cc:dd:ee:ff",
			"aa:bb:cc:dd:ee:11",
			"fake",
			"eth0",
			"bond0",
			false,
		},
		// This is a bond eg. bond0
		{
			"mac:aa:bb:*",
			"aa:bb:cc:dd:ee:11",
			"bond",
			"bond0",
			"bond0",
			false,
		},
	}

	for i, tt := range testCases {
		netCfg := NetworkConfig{
			Interfaces: map[string]InterfaceConfig{
				tt.name: InterfaceConfig{
					Match: tt.match,
					Bond:  tt.bond,
				},
			},
		}

		linkAttrs := netlink.NewLinkAttrs()
		linkAttrs.Name = tt.name
		linkAttrs.HardwareAddr, _ = net.ParseMAC(tt.mac)
		link := mockLink{attrs: linkAttrs}

		_, match := findMatch(link, &netCfg)

		if match != tt.expected {
			t.Errorf("Test case %d failed: mac: '%s' match '%s' expected: '%v' got: '%v'", i, tt.mac, tt.match, tt.expected, match)
		}
	}

}
