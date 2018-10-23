package aliyun

import (
	"fmt"
	"log"
	"strings"

	"github.com/rancher/os/config/cloudinit/datasource"
	"github.com/rancher/os/config/cloudinit/datasource/metadata"
	"github.com/rancher/os/pkg/netconf"
)

const (
	DefaultAddress = "http://100.100.100.200/"
	apiVersion     = "2016-01-01/"
	userdataPath   = apiVersion + "user-data/"
	metadataPath   = apiVersion + "meta-data/"
)

type MetadataService struct {
	metadata.Service
}

func NewDatasource(root string) *MetadataService {
	if root == "" {
		root = DefaultAddress
	}
	return &MetadataService{metadata.NewDatasource(root, apiVersion, userdataPath, metadataPath, nil)}
}

func (ms MetadataService) AvailabilityChanges() bool {
	// TODO: if it can't find the network, maybe we can start it?
	return false
}

func (ms MetadataService) FetchMetadata() (metadata datasource.Metadata, err error) {
	// see https://www.alibabacloud.com/help/faq-detail/49122.htm
	metadata.NetworkConfig = netconf.NetworkConfig{}

	enablePublicKey := false

	rootContents, err := ms.FetchAttributes("")
	if err != nil {
		return metadata, err
	}
	for _, c := range rootContents {
		if c == "public-keys/" {
			enablePublicKey = true
			break
		}
	}
	if !enablePublicKey {
		return metadata, fmt.Errorf("The public-keys should be enable in %s", ms.Type())
	}

	keynames, err := ms.FetchAttributes("public-keys/")
	if err != nil {
		return metadata, err
	}

	metadata.SSHPublicKeys = map[string]string{}
	for _, k := range keynames {
		k = strings.TrimRight(k, "/")
		sshkey, err := ms.FetchAttribute(fmt.Sprintf("public-keys/%s/openssh-key", k))
		if err != nil {
			return metadata, err
		}
		metadata.SSHPublicKeys[k] = sshkey
		log.Printf("Found SSH key for %q\n", k)
	}

	if hostname, err := ms.FetchAttribute("hostname"); err == nil {
		metadata.Hostname = hostname
		log.Printf("Found hostname  %s\n", hostname)
	} else {
		return metadata, err
	}

	return metadata, nil
}

func (ms MetadataService) Type() string {
	return "aliyun-metadata-service"
}
