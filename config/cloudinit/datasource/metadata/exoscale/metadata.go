package exoscale

import (
	"net"
	"strconv"
	"strings"

	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/datasource/metadata"
	"github.com/rancher/os/config/cloudinit/pkg"
	"github.com/rancher/os/pkg/log"
)

const (
	defaultAddress = "http://169.254.169.254/"
	apiVersion     = "1.0/"
	userdataPath   = apiVersion + "user-data"
	metadataPath   = apiVersion + "meta-data/"
)

type MetadataService struct {
	metadata.Service
}

func NewDatasource(root string) *MetadataService {
	if root == "" {
		root = defaultAddress
	}

	return &MetadataService{
		metadata.NewDatasourceWithCheckPath(
			root,
			apiVersion,
			metadataPath,
			userdataPath,
			metadataPath,
			nil,
		),
	}
}

func (ms MetadataService) AvailabilityChanges() bool {
	// TODO: if it can't find the network, maybe we can start it?
	return false
}

func (ms MetadataService) FetchMetadata() (datasource.Metadata, error) {
	metadata := datasource.Metadata{}

	if sshKeys, err := ms.FetchAttributes("public-keys"); err == nil {
		metadata.SSHPublicKeys = map[string]string{}
		for i, sshkey := range sshKeys {
			log.Printf("Found SSH key %d", i)
			metadata.SSHPublicKeys[strconv.Itoa(i)] = sshkey
		}
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if hostname, err := ms.FetchAttribute("local-hostname"); err == nil {
		metadata.Hostname = strings.Split(hostname, " ")[0]
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	if localAddr, err := ms.FetchAttribute("local-ipv4"); err == nil {
		metadata.PrivateIPv4 = net.ParseIP(localAddr)
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}
	if publicAddr, err := ms.FetchAttribute("public-ipv4"); err == nil {
		metadata.PublicIPv4 = net.ParseIP(publicAddr)
	} else if _, ok := err.(pkg.ErrNotFound); !ok {
		return metadata, err
	}

	return metadata, nil
}

func (ms MetadataService) Type() string {
	return "exoscale-metadata-service"
}
