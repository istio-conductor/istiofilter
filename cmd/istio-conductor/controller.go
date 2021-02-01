package main

import (
	"context"

	"github.com/istio-conductor/istiofilter/client-go/pkg/clientset/versioned/scheme"
	"github.com/istio-conductor/istiofilter/pkg/controller"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const leaderName = "istio-conductor-leader"

func startController(ctx context.Context, privilegeNamespaces []string) error {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Could not get apiserver config: %v\n", err)
		return err
	}
	mgrOpt := manager.Options{
		MetricsBindAddress:      "0",
		LeaderElection:          true,
		LeaderElectionNamespace: "istio-system",
		LeaderElectionID:        leaderName,
	}
	m, err := manager.New(cfg, mgrOpt)
	if err != nil {
		log.Fatalf("Could not create a controller manager: %v", err)
		return err
	}
	err = controller.AddIstioFilter(m, privilegeNamespaces)
	if err != nil {
		log.Fatalf("AddIstioFilter: %v", err)
		return err
	}
	err = scheme.AddToScheme(m.GetScheme())
	if err != nil {
		log.Fatalf("AddToScheme: %v", err)
	}
	return m.Start(ctx)
}
