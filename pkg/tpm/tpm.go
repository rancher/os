package tpm

import (
	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	roscontrollers "github.com/rancher/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	corecontrollers "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
)

const (
	machineByHash = "machineByHash"
	tpmCACert     = "tpm-ca"
)

type AuthServer struct {
	machineCache  roscontrollers.MachineInventoryCache
	machineClient roscontrollers.MachineInventoryClient
	secretCache   corecontrollers.SecretCache
}

func New(clients *clients.Clients) *AuthServer {
	a := &AuthServer{
		machineCache:  clients.OS.MachineInventory().Cache(),
		machineClient: clients.OS.MachineInventory(),
		secretCache:   clients.Core.Secret().Cache(),
	}

	a.machineCache.AddIndexer(machineByHash, func(obj *v1.MachineInventory) ([]string, error) {
		if obj.Spec.TPMHash == "" {
			return nil, nil
		}
		return []string{obj.Spec.TPMHash}, nil
	})

	return a
}
