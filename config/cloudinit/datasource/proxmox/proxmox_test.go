package proxmox

import (
	"testing"

	"github.com/rancher/os/config/cloudinit/datasource/test"
)

func TestFetchUserdata(t *testing.T) {
	for _, tt := range []struct {
		root     string
		files    test.MockFilesystem
		userdata string
	}{
		{
			root:     "/",
			files:    test.NewMockFilesystem(),
			userdata: "",
		},
		{
			root:     "/media/config",
			files:    test.NewMockFilesystem(test.File{Path: "/media/config/user-data", Contents: "userdata"}),
			userdata: "userdata",
		},
	} {
		pve := Proxmox{tt.root, tt.files.ReadFile, nil, true}
		userdata, err := pve.FetchUserdata()
		if err != nil {
			t.Fatalf("bad error for %+v: want %v, got %q", tt, nil, err)
		}
		if string(userdata) != tt.userdata {
			t.Fatalf("bad userdata for %+v: want %q, got %q", tt, tt.userdata, userdata)
		}
	}
}

func TestConfigRoot(t *testing.T) {
	for _, tt := range []struct {
		root       string
		configRoot string
	}{
		{
			root:       "/",
			configRoot: "/",
		},
		{
			root:       "/media/pve-config",
			configRoot: "/media/pve-config",
		},
	} {
		pve := Proxmox{tt.root, nil, nil, true}
		if configRoot := pve.ConfigRoot(); configRoot != tt.configRoot {
			t.Fatalf("bad config root for %q: want %q, got %q", tt, tt.configRoot, configRoot)
		}
	}
}

func TestNewDataSource(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectRoot string
	}{
		{
			root:       "",
			expectRoot: "",
		},
		{
			root:       "/media/pve-config",
			expectRoot: "/media/pve-config",
		},
	} {
		service := NewDataSource(tt.root)
		if service.root != tt.expectRoot {
			t.Fatalf("bad root (%q): want %q, got %q", tt.root, tt.expectRoot, service.root)
		}
	}
}
