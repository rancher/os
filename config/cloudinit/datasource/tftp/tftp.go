package tftp

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/burmilla/os/config/cloudinit/datasource"

	"github.com/pin/tftp"
)

type Client interface {
	Receive(filename string, mode string) (io.WriterTo, error)
}

type RemoteFile struct {
	host      string
	path      string
	client    Client
	stream    io.WriterTo
	lastError error
}

func NewDatasource(hostAndPath string) *RemoteFile {
	parts := strings.SplitN(hostAndPath, "/", 2)

	if len(parts) < 2 {
		return &RemoteFile{hostAndPath, "", nil, nil, nil}
	}

	host := parts[0]
	if match, _ := regexp.MatchString(":[0-9]{2,5}$", host); !match {
		// No port, using default port 69
		host += ":69"
	}

	path := parts[1]
	if client, lastError := tftp.NewClient(host); lastError == nil {
		return &RemoteFile{host, path, client, nil, nil}
	}

	return &RemoteFile{host, path, nil, nil, nil}
}

func (f *RemoteFile) IsAvailable() bool {
	f.stream, f.lastError = f.client.Receive(f.path, "octet")
	return f.lastError == nil
}

func (f *RemoteFile) Finish() error {
	return nil
}

func (f *RemoteFile) String() string {
	return fmt.Sprintf("%s, host:%s, path:%s (lastError: %v)", f.Type(), f.host, f.path, f.lastError)
}

func (f *RemoteFile) AvailabilityChanges() bool {
	return false
}

func (f *RemoteFile) ConfigRoot() string {
	return ""
}

func (f *RemoteFile) FetchMetadata() (datasource.Metadata, error) {
	return datasource.Metadata{}, nil
}

func (f *RemoteFile) FetchUserdata() ([]byte, error) {
	var b bytes.Buffer

	_, err := f.stream.WriteTo(&b)

	return b.Bytes(), err
}

func (f *RemoteFile) Type() string {
	return "tftp"
}
