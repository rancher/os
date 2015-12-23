package metadata

import (
	"net/http"
	"net/url"

	"github.com/packethost/packngo"
)

const (
	baseUrl = "https://metadata.packet.net"
)

type Client struct {
	client *packngo.Client

	Metadata MetadataService
	Userdata UserdataService
}

type MetadataService interface {
	Get() (Metadata, error)
}

type UserdataService interface {
	Get() (string, error)
}

func NewClient(httpClient *http.Client) *Client {
	c := packngo.NewClient("", "", httpClient)
	c.BaseURL, _ = url.Parse(baseUrl)
	return &Client{
		client:   c,
		Metadata: &MetadataServiceOp{client: c},
		Userdata: &UserdataServiceOp{client: c},
	}
}
