package inventory

import (
	"context"
	"fmt"

	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	ranchercontrollers "github.com/rancher/os2/pkg/generated/controllers/management.cattle.io/v3"
	provcontrollers "github.com/rancher/os2/pkg/generated/controllers/provisioning.cattle.io/v1"
	v12 "github.com/rancher/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/name"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type handler struct {
	clusterCache                   provcontrollers.ClusterCache
	clusterRegistrationTokenCache  ranchercontrollers.ClusterRegistrationTokenCache
	clusterRegistrationTokenClient ranchercontrollers.ClusterRegistrationTokenClient
}

func Register(ctx context.Context, clients *clients.Clients) {
	h := &handler{
		clusterCache:                   clients.Provisioning.Cluster().Cache(),
		clusterRegistrationTokenCache:  clients.Rancher.ClusterRegistrationToken().Cache(),
		clusterRegistrationTokenClient: clients.Rancher.ClusterRegistrationToken(),
	}

	clients.OS.MachineInventory().OnRemove(ctx, "machine-inventory-remove", h.OnMachineInventoryRemove)
	v12.RegisterMachineInventoryStatusHandler(ctx, clients.OS.MachineInventory(), "", "machine-inventory", h.OnMachineInventory)
}

func (h *handler) OnMachineInventoryRemove(key string, machine *v1.MachineInventory) (*v1.MachineInventory, error) {
	if machine.Status.ClusterRegistrationTokenName != "" && machine.Status.ClusterRegistrationTokenNamespace != "" {
		err := h.clusterRegistrationTokenClient.Delete(machine.Status.ClusterRegistrationTokenNamespace, machine.Status.ClusterRegistrationTokenName, nil)
		if !apierrors.IsNotFound(err) && err != nil {
			return nil, err
		}
	}
	return machine, nil
}

func (h *handler) OnMachineInventory(machine *v1.MachineInventory, status v1.MachineInventoryStatus) (v1.MachineInventoryStatus, error) {
	if machine == nil || machine.Spec.ClusterName == "" {
		return status, nil
	}

	cluster, err := h.clusterCache.Get(machine.Namespace, machine.Spec.ClusterName)
	if err != nil {
		return status, err
	}

	if cluster.Status.ClusterName == "" {
		return status, fmt.Errorf("waiting for mgmt cluster to be created for prov cluster %s/%s", machine.Namespace, machine.Spec.ClusterName)
	}

	crtName := name.SafeConcatName(cluster.Status.ClusterName, machine.Name, "token")
	_, err = h.clusterRegistrationTokenCache.Get(cluster.Status.ClusterName, crtName)
	if apierrors.IsNotFound(err) {
		_, err = h.clusterRegistrationTokenClient.Create(&v3.ClusterRegistrationToken{
			ObjectMeta: metav1.ObjectMeta{
				Name:      crtName,
				Namespace: cluster.Status.ClusterName,
			},
			Spec: v3.ClusterRegistrationTokenSpec{
				ClusterName: cluster.Status.ClusterName,
			},
		})
	}

	if err != nil {
		return status, err
	}

	status.ClusterRegistrationTokenName = crtName
	status.ClusterRegistrationTokenNamespace = cluster.Status.ClusterName
	return status, nil
}
