package main

import (
	"context"
	"flag"
	"os"

	"github.com/rancher/os/pkg/operator"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
)

var (
	namespace = flag.String("namespace", "cattle-rancheros-operator-system", "Namespace of the pod")
)

func main() {
	flag.Parse()
	logrus.Info("Starting controller")
	ctx := signals.SetupSignalHandler(context.Background())

	if os.Getenv("NAMESPACE") != "" {
		*namespace = os.Getenv("NAMESPACE")
	}
	if err := operator.Run(ctx, *namespace); err != nil {
		logrus.Fatalf("Error starting: %s", err.Error())
	}

	<-ctx.Done()
}
