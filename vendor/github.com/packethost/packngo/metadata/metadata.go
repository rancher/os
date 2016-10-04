package metadata

import (
	"github.com/packethost/packngo"
)

const (
	metadataBasePath = "/metadata"
)

type Metadata struct {
	PhoneHomeURL    string          `json:"phone_home_url"`
	ApiUrl          string          `json:"api_url"`
	Id              string          `json:"id"`
	Hostname        string          `json:"hostname"`
	Iqn             string          `json:"iqn"`
	OperatingSystem OperatingSystem `json:"operating_system"`
	Plan            string          `json:"plan"`
	Facility        string          `json:"facility"`
	SshKeys         []string        `json:"ssh_keys"`
	Network         Network         `json:"network"`
}

type Network struct {
	Addresses  []Address   `json:"addresses"`
	Interfaces []Interface `json:"interfaces"`
}

type Address struct {
	Href          string    `json:"href"`
	Gateway       string    `json:"gateway"`
	Address       string    `json:"address"`
	Network       string    `json:"network"`
	Id            string    `json:"id"`
	AddressFamily int       `json:"address_family"`
	Netmask       string    `json:"netmask"`
	Public        bool      `json:"public"`
	Cidr          int       `json:"cidr"`
	Management    bool      `json:"management"`
	Manageable    bool      `json:"manageable"`
	AssignedTo    Reference `json:"assigned_to"`
}

type Reference struct {
	Href string `json:"href"`
}

type OperatingSystem struct {
	Version string `json:"version"`
	Distro  string `json:"distro"`
	Slug    string `json:"slug"`
}

type Interface struct {
	Mac  string `json:"mac"`
	Name string `json:"name"`
}

type MetadataServiceOp struct {
	client *packngo.Client
}

func (s *MetadataServiceOp) Get() (Metadata, error) {
	metadata := Metadata{}

	req, err := s.client.NewRequest("GET", metadataBasePath, nil)
	if err != nil {
		return metadata, err
	}

	_, err = s.client.Do(req, &metadata)
	return metadata, err
}
