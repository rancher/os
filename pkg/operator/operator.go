package operator

import (
	"context"

	v1 "github.com/rancher/os2/pkg/apis/rancheros.cattle.io/v1"
	"github.com/rancher/os2/pkg/clients"
	"github.com/rancher/os2/pkg/controllers/inventory"
	"github.com/rancher/os2/pkg/controllers/managedos"
	"github.com/rancher/os2/pkg/server"
	"github.com/rancher/steve/pkg/aggregation"
	"github.com/rancher/wrangler/pkg/crd"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func Run(ctx context.Context, namespace string) error {
	restConfig, err := config.GetConfig()
	if err != nil {
		logrus.Fatalf("failed to find kubeconfig: %v", err)
	}

	clients, err := clients.NewFromConfig(restConfig)
	if err != nil {
		logrus.Fatalf("Error building controller: %s", err.Error())
	}

	factory, err := crd.NewFactoryFromClient(restConfig)
	if err != nil {
		logrus.Fatalf("Failed to create CRD factory: %v", err)
	}

	err = factory.BatchCreateCRDs(ctx,
		crd.CRD{
			SchemaObject: v1.ManagedOSImage{},
			Status:       true,
		},
		crd.CRD{
			SchemaObject: v1.MachineInventory{},
			Status:       true,
		},
	).BatchWait()
	if err != nil {
		logrus.Fatalf("Failed to create CRDs: %v", err)
	}

	managedos.Register(ctx, clients)
	inventory.Register(ctx, clients)

	aggregation.Watch(ctx, clients.Core.Secret(), namespace, "rancheros-operator", server.New(clients))
	return clients.Start(ctx)
}
