// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/coreos/coreos-cloudinit/config"
)

func TestPlaceUnit(t *testing.T) {
	tests := []config.Unit{
		{
			Name:    "50-eth0.network",
			Runtime: true,
			Content: "[Match]\nName=eth47\n\n[Network]\nAddress=10.209.171.177/19\n",
		},
		{
			Name:    "media-state.mount",
			Content: "[Mount]\nWhat=/dev/sdb1\nWhere=/media/state\n",
		},
	}

	for _, tt := range tests {
		dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
		if err != nil {
			panic(fmt.Sprintf("Unable to create tempdir: %v", err))
		}

		u := Unit{tt}
		sd := &systemd{dir}

		if err := sd.PlaceUnit(u); err != nil {
			t.Fatalf("PlaceUnit(): bad error (%+v): want nil, got %s", tt, err)
		}

		fi, err := os.Stat(u.Destination(dir))
		if err != nil {
			t.Fatalf("Stat(): bad error (%+v): want nil, got %s", tt, err)
		}

		if mode := fi.Mode(); mode != os.FileMode(0644) {
			t.Errorf("bad filemode (%+v): want %v, got %v", tt, os.FileMode(0644), mode)
		}

		c, err := ioutil.ReadFile(u.Destination(dir))
		if err != nil {
			t.Fatalf("ReadFile(): bad error (%+v): want nil, got %s", tt, err)
		}

		if string(c) != tt.Content {
			t.Errorf("bad contents (%+v): want %q, got %q", tt, tt.Content, string(c))
		}

		os.RemoveAll(dir)
	}
}

func TestPlaceUnitDropIn(t *testing.T) {
	tests := []config.Unit{
		{
			Name:    "false.service",
			Runtime: true,
			DropIns: []config.UnitDropIn{
				{
					Name:    "00-true.conf",
					Content: "[Service]\nExecStart=\nExecStart=/usr/bin/true\n",
				},
			},
		},
		{
			Name: "true.service",
			DropIns: []config.UnitDropIn{
				{
					Name:    "00-false.conf",
					Content: "[Service]\nExecStart=\nExecStart=/usr/bin/false\n",
				},
			},
		},
	}

	for _, tt := range tests {
		dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
		if err != nil {
			panic(fmt.Sprintf("Unable to create tempdir: %v", err))
		}

		u := Unit{tt}
		sd := &systemd{dir}

		if err := sd.PlaceUnitDropIn(u, u.DropIns[0]); err != nil {
			t.Fatalf("PlaceUnit(): bad error (%+v): want nil, got %s", tt, err)
		}

		fi, err := os.Stat(u.DropInDestination(dir, u.DropIns[0]))
		if err != nil {
			t.Fatalf("Stat(): bad error (%+v): want nil, got %s", tt, err)
		}

		if mode := fi.Mode(); mode != os.FileMode(0644) {
			t.Errorf("bad filemode (%+v): want %v, got %v", tt, os.FileMode(0644), mode)
		}

		c, err := ioutil.ReadFile(u.DropInDestination(dir, u.DropIns[0]))
		if err != nil {
			t.Fatalf("ReadFile(): bad error (%+v): want nil, got %s", tt, err)
		}

		if string(c) != u.DropIns[0].Content {
			t.Errorf("bad contents (%+v): want %q, got %q", tt, u.DropIns[0].Content, string(c))
		}

		os.RemoveAll(dir)
	}
}

func TestMachineID(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	os.Mkdir(path.Join(dir, "etc"), os.FileMode(0755))
	ioutil.WriteFile(path.Join(dir, "etc", "machine-id"), []byte("node007\n"), os.FileMode(0444))

	if MachineID(dir) != "node007" {
		t.Fatalf("File has incorrect contents")
	}
}

func TestMaskUnit(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	sd := &systemd{dir}

	// Ensure mask works with units that do not currently exist
	uf := Unit{config.Unit{Name: "foo.service"}}
	if err := sd.MaskUnit(uf); err != nil {
		t.Fatalf("Unable to mask new unit: %v", err)
	}
	fooPath := path.Join(dir, "etc", "systemd", "system", "foo.service")
	fooTgt, err := os.Readlink(fooPath)
	if err != nil {
		t.Fatal("Unable to read link", err)
	}
	if fooTgt != "/dev/null" {
		t.Fatal("unit not masked, got unit target", fooTgt)
	}

	// Ensure mask works with unit files that already exist
	ub := Unit{config.Unit{Name: "bar.service"}}
	barPath := path.Join(dir, "etc", "systemd", "system", "bar.service")
	if _, err := os.Create(barPath); err != nil {
		t.Fatalf("Error creating new unit file: %v", err)
	}
	if err := sd.MaskUnit(ub); err != nil {
		t.Fatalf("Unable to mask existing unit: %v", err)
	}
	barTgt, err := os.Readlink(barPath)
	if err != nil {
		t.Fatal("Unable to read link", err)
	}
	if barTgt != "/dev/null" {
		t.Fatal("unit not masked, got unit target", barTgt)
	}
}

func TestUnmaskUnit(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	sd := &systemd{dir}

	nilUnit := Unit{config.Unit{Name: "null.service"}}
	if err := sd.UnmaskUnit(nilUnit); err != nil {
		t.Errorf("unexpected error from unmasking nonexistent unit: %v", err)
	}

	uf := Unit{config.Unit{Name: "foo.service", Content: "[Service]\nExecStart=/bin/true"}}
	dst := uf.Destination(dir)
	if err := os.MkdirAll(path.Dir(dst), os.FileMode(0755)); err != nil {
		t.Fatalf("Unable to create unit directory: %v", err)
	}
	if _, err := os.Create(dst); err != nil {
		t.Fatalf("Unable to write unit file: %v", err)
	}

	if err := ioutil.WriteFile(dst, []byte(uf.Content), 700); err != nil {
		t.Fatalf("Unable to write unit file: %v", err)
	}
	if err := sd.UnmaskUnit(uf); err != nil {
		t.Errorf("unmask of non-empty unit returned unexpected error: %v", err)
	}
	got, _ := ioutil.ReadFile(dst)
	if string(got) != uf.Content {
		t.Errorf("unmask of non-empty unit mutated unit contents unexpectedly")
	}

	ub := Unit{config.Unit{Name: "bar.service"}}
	dst = ub.Destination(dir)
	if err := os.Symlink("/dev/null", dst); err != nil {
		t.Fatalf("Unable to create masked unit: %v", err)
	}
	if err := sd.UnmaskUnit(ub); err != nil {
		t.Errorf("unmask of unit returned unexpected error: %v", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("expected %s to not exist after unmask, but got err: %s", dst, err)
	}
}

func TestNullOrEmpty(t *testing.T) {
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer os.RemoveAll(dir)

	non := path.Join(dir, "does_not_exist")
	ne, err := nullOrEmpty(non)
	if !os.IsNotExist(err) {
		t.Errorf("nullOrEmpty on nonexistent file returned bad error: %v", err)
	}
	if ne {
		t.Errorf("nullOrEmpty returned true unxpectedly")
	}

	regEmpty := path.Join(dir, "regular_empty_file")
	_, err = os.Create(regEmpty)
	if err != nil {
		t.Fatalf("Unable to create tempfile: %v", err)
	}
	gotNe, gotErr := nullOrEmpty(regEmpty)
	if !gotNe || gotErr != nil {
		t.Errorf("nullOrEmpty of regular empty file returned %t, %v - want true, nil", gotNe, gotErr)
	}

	reg := path.Join(dir, "regular_file")
	if err := ioutil.WriteFile(reg, []byte("asdf"), 700); err != nil {
		t.Fatalf("Unable to create tempfile: %v", err)
	}
	gotNe, gotErr = nullOrEmpty(reg)
	if gotNe || gotErr != nil {
		t.Errorf("nullOrEmpty of regular file returned %t, %v - want false, nil", gotNe, gotErr)
	}

	null := path.Join(dir, "null")
	if err := os.Symlink(os.DevNull, null); err != nil {
		t.Fatalf("Unable to create /dev/null link: %s", err)
	}
	gotNe, gotErr = nullOrEmpty(null)
	if !gotNe || gotErr != nil {
		t.Errorf("nullOrEmpty of null symlink returned %t, %v - want true, nil", gotNe, gotErr)
	}

}
