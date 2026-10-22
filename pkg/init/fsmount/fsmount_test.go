package fsmount

import "testing"

func TestDiscardOptions(t *testing.T) {
	for fsType, want := range map[string]string{
		"ext4":  "discard",
		"xfs":   "discard",
		"btrfs": "discard",
		"f2fs":  "discard",
		"vfat":  "",
		"":      "",
	} {
		if got := discardOptions(fsType); got != want {
			t.Errorf("discardOptions(%q) = %q, want %q", fsType, got, want)
		}
	}
}
