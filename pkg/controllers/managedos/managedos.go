package managedos

import (
	"context"
	"fmt"

	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	provv1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	fleetcontrollers "github.com/rancher/os2/pkg/generated/controllers/fleet.cattle.io/v1alpha1"
	ranchercontrollers "github.com/rancher/os2/pkg/generated/controllers/management.cattle.io/v3"
	oscontrollers "github.com/rancher/os2/pkg/generated/controllers/rancheros.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/name"
	"github.com/rancher/wrangler/pkg/relatedresource"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Register(ctx context.Context, clients *clients.Clients) {
	h := &handler{
		bundleCache:   clients.Fleet.Bundle().Cache(),
		settingsCache: clients.Rancher.Setting().Cache(),
	}

	relatedresource.Watch(ctx,
		"mcc-from-bundle-trigger",
		relatedresource.OwnerResolver(true, provv1.SchemeGroupVersion.String(), "ManagedOSImage"),
		clients.OS.ManagedOSImage(),
		clients.Fleet.Bundle())
	oscontrollers.RegisterManagedOSImageGeneratingHandler(ctx,
		clients.OS.ManagedOSImage(),
		clients.Apply.
			WithSetOwnerReference(true, true).
			WithCacheTypes(
				clients.OS.ManagedOSImage(),
				clients.Fleet.Bundle()),
		"Defined",
		"mos-bundle",
		h.OnChange,
		nil)
}

type handler struct {
	bundleCache   fleetcontrollers.BundleCache
	settingsCache ranchercontrollers.SettingCache
}

func (h *handler) defaultRegistry() (string, error) {
	setting, err := h.settingsCache.Get("system-default-registry")
	if err != nil {
		return "", err
	}
	if setting.Value == "" {
		return setting.Default, nil
	}
	return setting.Value, nil
}

func (h *handler) OnChange(mos *provv1.ManagedOSImage, status provv1.ManagedOSImageStatus) ([]runtime.Object, provv1.ManagedOSImageStatus, error) {
	if mos.Spec.OSImage == "" {
		return nil, status, nil
	}

	prefix, err := h.defaultRegistry()
	if err != nil {
		return nil, status, err
	}

	objs, err := objects(mos, prefix)
	if err != nil {
		return nil, status, err
	}

	resources, err := ToResources(objs)
	if err != nil {
		return nil, status, err
	}

	if mos.Namespace == "fleet-local" && len(mos.Spec.Targets) > 0 {
		return nil, status, fmt.Errorf("spec.targets should be empty if in the fleet-local namespace")
	}

	bundle := &v1alpha1.Bundle{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name.SafeConcatName("mos", mos.Name),
			Namespace: mos.Namespace,
		},
		Spec: v1alpha1.BundleSpec{
			Resources:               resources,
			BundleDeploymentOptions: v1alpha1.BundleDeploymentOptions{},
			RolloutStrategy:         mos.Spec.ClusterRolloutStrategy,
			Targets:                 mos.Spec.Targets,
		},
	}

	if mos.Namespace == "fleet-local" {
		bundle.Spec.Targets = []v1alpha1.BundleTarget{{ClusterName: "local"}}
	}

	status, err = h.updateStatus(status, bundle)
	return []runtime.Object{
		bundle,
	}, status, err
}

func (h *handler) updateStatus(status provv1.ManagedOSImageStatus, bundle *v1alpha1.Bundle) (provv1.ManagedOSImageStatus, error) {
	bundle, err := h.bundleCache.Get(bundle.Namespace, bundle.Name)
	if apierrors.IsNotFound(err) {
		return status, nil
	} else if err != nil {
		return status, err
	}

	status.BundleStatus = bundle.Status
	return status, nil
}
