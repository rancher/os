package main

import (
	"os"

	"github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	v3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	provv1 "github.com/rancher/rancher/pkg/apis/provisioning.cattle.io/v1"
	controllergen "github.com/rancher/wrangler/pkg/controller-gen"
	"github.com/rancher/wrangler/pkg/controller-gen/args"
)

func main() {
	os.Unsetenv("GOPATH")
	controllergen.Run(args.Options{
		OutputPackage: "github.com/rancher/os/pkg/generated",
		Boilerplate:   "scripts/boilerplate.go.txt",
		Groups: map[string]args.Group{
			"provisioning.cattle.io": {
				Types: []interface{}{
					provv1.Cluster{},
				},
			},
			"management.cattle.io": {
				Types: []interface{}{
					v3.Setting{},
					v3.ClusterRegistrationToken{},
				},
			},
			"fleet.cattle.io": {
				Types: []interface{}{
					v1alpha1.Bundle{},
				},
			},
			"rancheros.cattle.io": {
				Types: []interface{}{
					"./pkg/apis/rancheros.cattle.io/v1",
				},
				GenerateTypes: true,
			},
		},
	})
}
