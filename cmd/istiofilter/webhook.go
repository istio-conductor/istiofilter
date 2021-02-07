package main

import (
	"context"

	"github.com/istio-conductor/istiofilter/pkg/patch"
	"github.com/istio-conductor/istiofilter/pkg/store/informer"
	"github.com/istio-conductor/istiofilter/webhook"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Config struct {
	Port                int      `json:"port"`
	CertFile            string   `json:"certFile"`
	KeyFile             string   `json:"keyFile"`
	PrivilegeNamespaces []string `json:"privilegeNamespaces"`
}

func Serve(ctx context.Context, c *Config) error {
	logrus.Debug("begin webhook serve")
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}
	logrus.Debug("create informer store")
	store, err := informer.New(c.PrivilegeNamespaces, cfg)
	if err != nil {
		return err
	}
	store.Start(ctx.Done())
	logrus.Debug("create patcher")
	patcher := patch.New(store)
	logrus.Debug("webhook running")
	return webhook.Run(ctx, c.Port, c.CertFile, c.KeyFile, patcher)
}
