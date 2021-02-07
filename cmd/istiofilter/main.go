package main

import (
	"os"
	"strings"

	"github.com/pascaldekloe/name"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/sync/errgroup"
	"istio.io/istio/pkg/cmd"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var (
	startHook = true
	logLevel  = "info"
)
var (
	rootCmd = &cobra.Command{
		Use:          "istiofilter",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.SetReportCaller(true)
			level, _ := logrus.ParseLevel(logLevel)
			logrus.SetLevel(level)
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			cmd.PrintFlags(c.Flags())

			ctx := signals.SetupSignalHandler()

			group, ctx := errgroup.WithContext(ctx)
			group.Go(func() error {
				return startController(ctx, webhookCfg.PrivilegeNamespaces)
			})
			if startHook {
				group.Go(func() error {
					return Serve(ctx, webhookCfg)
				})
			}
			err := group.Wait()
			return err
		},
	}
)
var webhookCfg = &Config{}

func init() {
	cobra.OnInitialize(func() {
		BindEnv(rootCmd)
	})
	rootCmd.PersistentFlags().StringVar(&logLevel, "logLevel", "info", "")
	rootCmd.PersistentFlags().BoolVar(&startHook, "starthook", true, "")
	rootCmd.PersistentFlags().StringVarP(
		&webhookCfg.KeyFile, "keyFile", "k", "", "")
	rootCmd.PersistentFlags().StringVarP(
		&webhookCfg.CertFile, "certFile", "c", "", "")
	rootCmd.PersistentFlags().IntVarP(
		&webhookCfg.Port, "port", "p", 443, "")
	rootCmd.PersistentFlags().StringSliceVarP(
		&webhookCfg.PrivilegeNamespaces, "privilegeNamespaces", "n", []string{"istio-system"}, "")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
}

func BindEnv(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flag.Changed = true
		if !flag.Changed {
			env, ok := os.LookupEnv(strings.ToUpper(name.Delimit(flag.Name, '_')))
			if ok {
				_ = cmd.Flags().Set(flag.Name, env)
			}
		}
	})
}
