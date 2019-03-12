package tftp

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

type mockClient struct {
}

type mockReceiver struct {
}

func (r mockReceiver) WriteTo(w io.Writer) (n int64, err error) {
	b := []byte("cloud-config file")
	w.Write(b)
	return int64(len(b)), nil
}

func (c mockClient) Receive(filename string, mode string) (io.WriterTo, error) {
	if filename == "does-not-exist" {
		return &mockReceiver{}, fmt.Errorf("does not exist")
	}
	return &mockReceiver{}, nil
}

var _ Client = (*mockClient)(nil)

func TestNewDatasource(t *testing.T) {
	for _, tt := range []struct {
		root       string
		expectHost string
		expectPath string
	}{
		{
			root:       "127.0.0.1/test/file.yaml",
			expectHost: "127.0.0.1:69",
			expectPath: "test/file.yaml",
		},
		{
			root:       "127.0.0.1/test/file.yaml",
			expectHost: "127.0.0.1:69",
			expectPath: "test/file.yaml",
		},
	} {
		ds := NewDatasource(tt.root)
		if ds.host != tt.expectHost || ds.path != tt.expectPath {
			t.Fatalf("bad host or path (%q): want host=%s, got %s, path=%s, got %s", tt.root, tt.expectHost, ds.host, tt.expectPath, ds.path)
		}
	}
}

func TestIsAvailable(t *testing.T) {
	for _, tt := range []struct {
		remoteFile *RemoteFile
		expect     bool
	}{
		{
			remoteFile: &RemoteFile{"1.2.3.4", "test", &mockClient{}, nil, nil},
			expect:     true,
		},
		{
			remoteFile: &RemoteFile{"1.2.3.4", "does-not-exist", &mockClient{}, nil, nil},
			expect:     false,
		},
	} {
		if tt.remoteFile.IsAvailable() != tt.expect {
			t.Fatalf("expected remote file %s to be %v", tt.remoteFile.path, tt.expect)
		}
	}
}

func TestFetchUserdata(t *testing.T) {
	rf := &RemoteFile{"1.2.3.4", "test", &mockClient{}, &mockReceiver{}, nil}
	b, _ := rf.FetchUserdata()

	expect := []byte("cloud-config file")

	if len(b) != len(expect) || !reflect.DeepEqual(b, expect) {
		t.Fatalf("expected length of buffer to be %d was %d. Expected %s, got %s", len(expect), len(b), string(expect), string(b))
	}
}

func TestType(t *testing.T) {
	rf := &RemoteFile{"1.2.3.4", "test", &mockClient{}, nil, nil}

	if rf.Type() != "tftp" {
		t.Fatalf("expected remote file Type() to return %s got %s", "tftp", rf.Type())
	}
}
