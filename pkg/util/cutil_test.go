//go:build linux
// +build linux

package util

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGetFsType(t *testing.T) {
	mkfs, err := exec.LookPath("mkfs.ext4")
	if err != nil {
		t.Skip("mkfs.ext4 not available")
	}
	img := filepath.Join(t.TempDir(), "state.img")
	if err := os.WriteFile(img, make([]byte, 8<<20), 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(mkfs, "-q", "-L", "RANCHER_STATE", img).CombinedOutput(); err != nil {
		t.Fatalf("mkfs.ext4: %v: %s", err, out)
	}
	if got, err := GetFsType(img); err != nil || got != "ext4" {
		t.Errorf("GetFsType = %q, %v; want ext4", got, err)
	}
	if _, err := GetFsType(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Error("GetFsType on a missing file should fail")
	}
}
