package command

import (
	"log/slog"
	"os"
	"strings"

	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/upload"
	"github.com/jeffry-luqman/zlog"
	"github.com/spf13/viper"
)

func setupLogger() error {
	var (
		lvl slog.LevelVar
	)

	switch strings.ToLower(viper.GetString("log.level")) {
	case "panic":
		lvl.Set(slog.LevelError)
	case "fatal":
		lvl.Set(slog.LevelError)
	case "error":
		lvl.Set(slog.LevelError)
	case "warn":
		lvl.Set(slog.LevelWarn)
	case "info":
		lvl.Set(slog.LevelInfo)
	case "debug":
		lvl.Set(slog.LevelDebug)
	case "trace":
		lvl.Set(slog.LevelDebug)
	default:
		lvl.Set(slog.LevelInfo)
	}

	if viper.GetBool("log.pretty") {
		zlog.HandlerOptions = &slog.HandlerOptions{
			Level: &lvl,
		}

		slog.SetDefault(
			zlog.New(),
		)
	} else {
		slog.SetDefault(
			slog.New(
				slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
					Level: &lvl,
				}),
			),
		)
	}

	return nil
}

func setupConfig() {
	if viper.GetString("config.file") != "" {
		viper.SetConfigFile(viper.GetString("config.file"))
	} else {
		viper.SetConfigName("server")
		viper.AddConfigPath("/etc/gexec")
		viper.AddConfigPath("$HOME/.gexec")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("gexec")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := readConfig(); err != nil {
		slog.Error(
			"Failed to read config file",
			slog.Any("error", err),
		)
	}

	if err := viper.Unmarshal(cfg); err != nil {
		slog.Error(
			"Failed to parse config file",
			slog.Any("error", err),
		)
	}

	{
		passphrase, err := config.Value(cfg.Encrypt.Passphrase)

		if err != nil {
			slog.Error(
				"Failed to parse encrypt passphrase secret",
				slog.Any("error", err),
			)

			os.Exit(1)
		}

		if len(passphrase) != 32 {
			slog.Error(
				"Encryption passphrase got to be 32 chars",
				slog.Any("error", err),
			)

			os.Exit(1)
		}

		cfg.Encrypt.Passphrase = passphrase
	}
}

func readConfig() error {
	err := viper.ReadInConfig()

	if err == nil {
		return nil
	}

	if _, ok := err.(viper.ConfigFileNotFoundError); ok {
		return nil
	}

	if _, ok := err.(*os.PathError); ok {
		return nil
	}

	return err
}

func setupUploads(cfg *config.Config) (upload.Upload, error) {
	switch cfg.Upload.Driver {
	case "file":
		return upload.NewFileUpload(cfg.Upload)
	case "s3":
		return upload.NewS3Upload(cfg.Upload)
	case "minio":
		return upload.NewS3Upload(cfg.Upload)
	}

	return nil, upload.ErrUnknownDriver
}
