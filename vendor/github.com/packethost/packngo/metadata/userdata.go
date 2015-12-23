package metadata

import (
	"bytes"

	"github.com/packethost/packngo"
)

const (
	userdataBasePath = "/userdata"
)

type UserdataServiceOp struct {
	client *packngo.Client
}

func (s *UserdataServiceOp) Get() (string, error) {
	req, err := s.client.NewRequest("GET", userdataBasePath, nil)
	if err != nil {
		return "", err
	}

	buffer := &bytes.Buffer{}
	_, err = s.client.Do(req, buffer)
	return buffer.String(), err
}
