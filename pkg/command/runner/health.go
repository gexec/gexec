package command

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	healthCmd = &cobra.Command{
		Use:   "health",
		Short: "Perform health checks",
		Run:   healthAction,
		Args:  cobra.NoArgs,
	}
)

func init() {
	rootCmd.AddCommand(healthCmd)

	healthCmd.PersistentFlags().String("metrics-addr", defaultMetricsAddr, "Address to bind the metrics")
	viper.SetDefault("metrics.addr", defaultMetricsAddr)
	_ = viper.BindPFlag("metrics.addr", healthCmd.PersistentFlags().Lookup("metrics-addr"))
}

func healthAction(_ *cobra.Command, _ []string) {
	resp, err := http.Get(
		fmt.Sprintf(
			"http://%s/healthz",
			cfg.Metrics.Addr,
		),
	)

	if err != nil {
		slog.Error(
			"Failed to request health check",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		slog.Error(
			"Health seems to be in bad state",
			slog.Int("code", resp.StatusCode),
		)

		os.Exit(1)
	}

	slog.Debug(
		"Health check seems to be fine",
		slog.Int("code", resp.StatusCode),
	)
}
