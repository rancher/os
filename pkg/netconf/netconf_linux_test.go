package netconf

import (
	"net"
	"testing"

	"github.com/vishvananda/netlink"
)

type mockLink struct {
	attrs netlink.LinkAttrs
}

func (l mockLink) Attrs() *netlink.LinkAttrs {
	return &l.attrs
}

func (l mockLink) Type() string {
	return "fake"
}

func TestFindMatch(t *testing.T) {
	testCases := []struct {
		match    string
		mac      string
		expected bool
	}{
		{
			"mac:aa:bb:cc:dd:ee:ff",
			"aa:bb:cc:dd:ee:ff",
			true,
		},
		{
			"mac:aa:bb:cc:*",
			"aa:bb:cc:12:34:56",
			true,
		},
		{
			"mac:aa:bb:cc:*",
			"11:bb:cc:dd:ee:ff",
			false,
		},
		{
			"mac:aa:bb:cc:dd:ee:ff",
			"aa:bb:cc:dd:ee:11",
			false,
		},
	}

	for i, tt := range testCases {
		netCfg := NetworkConfig{
			Interfaces: map[string]InterfaceConfig{
				"eth0": InterfaceConfig{
					Match: tt.match,
				},
			},
		}

		linkAttrs := netlink.NewLinkAttrs()
		linkAttrs.Name = "eth0"
		linkAttrs.HardwareAddr, _ = net.ParseMAC(tt.mac)
		link := mockLink{attrs: linkAttrs}

		_, match := findMatch(link, &netCfg)

		if match != tt.expected {
			t.Errorf("Test case %d failed: mac: '%s' match '%s' expected: '%v' got: '%v'", i, tt.mac, tt.match, tt.expected, match)
		}
	}
}
