package command

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/metrics"
	"github.com/gexec/gexec/pkg/router"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	serverCmd = &cobra.Command{
		Use:   "start",
		Short: "Start integrated server",
		Run:   serverAction,
		Args:  cobra.NoArgs,
	}

	defaultMetricsAddr  = "0.0.0.0:8001"
	defaultMetricsToken = ""
	defaultMetricsPprof = false
	defaultRunnerServer = "http://localhost:8080"
	defaultRunnerToken  = ""
)

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.PersistentFlags().String("metrics-addr", defaultMetricsAddr, "Address to bind the metrics")
	viper.SetDefault("metrics.addr", defaultMetricsAddr)
	_ = viper.BindPFlag("metrics.addr", serverCmd.PersistentFlags().Lookup("metrics-addr"))

	serverCmd.PersistentFlags().String("metrics-token", defaultMetricsToken, "Token to make metrics secure")
	viper.SetDefault("metrics.token", defaultMetricsToken)
	_ = viper.BindPFlag("metrics.token", serverCmd.PersistentFlags().Lookup("metrics-token"))

	serverCmd.PersistentFlags().Bool("metrics-pprof", defaultMetricsPprof, "Enable pprof debugging")
	viper.SetDefault("metrics.pprof", defaultMetricsPprof)
	_ = viper.BindPFlag("metrics.pprof", serverCmd.PersistentFlags().Lookup("metrics-pprof"))

	serverCmd.PersistentFlags().String("runner-server", defaultRunnerServer, "Access to the server instance")
	viper.SetDefault("runner.server", defaultRunnerServer)
	_ = viper.BindPFlag("runner.server", serverCmd.PersistentFlags().Lookup("runner-server"))

	serverCmd.PersistentFlags().String("runner-token", defaultRunnerToken, "Access token for runner on server")
	viper.SetDefault("runner.token", defaultRunnerToken)
	_ = viper.BindPFlag("runner.token", serverCmd.PersistentFlags().Lookup("runner-token"))
}

func serverAction(_ *cobra.Command, _ []string) {
	token, err := config.Value(cfg.Metrics.Token)

	if err != nil {
		slog.Error(
			"Failed to parse metrics token secret",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	registry := metrics.New(
		metrics.WithNamespace("gexec_runner"),
		metrics.WithToken(token),
	)

	gr := run.Group{}

	{
		server := &http.Server{
			Addr: cfg.Metrics.Addr,
			Handler: router.Metrics(
				cfg,
				registry,
			),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		}

		gr.Add(func() error {
			slog.Info(
				"Starting metrics server",
				slog.String("addr", cfg.Metrics.Addr),
			)

			return server.ListenAndServe()
		}, func(reason error) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				slog.Error(
					"Failed to shutdown metrics gracefully",
					slog.Any("error", err),
				)

				return
			}

			slog.Info(
				"Metrics shutdown gracefully",
				slog.Any("reason", reason),
			)
		})
	}

	{
		stop := make(chan os.Signal, 1)

		gr.Add(func() error {
			signal.Notify(stop, os.Interrupt)

			<-stop

			return nil
		}, func(_ error) {
			close(stop)
		})
	}

	if err := gr.Run(); err != nil {
		os.Exit(1)
	}
}
