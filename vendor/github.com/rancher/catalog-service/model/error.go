package model

import "github.com/rancher/go-rancher/v2"

type CatalogError struct {
	client.Resource
	Status  string `json:"status"`
	Message string `json:"message"`
}
