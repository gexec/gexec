package command

import (
	"os"
	"strings"

	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/upload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func setupLogger() error {
	switch strings.ToLower(viper.GetString("log.level")) {
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if viper.GetBool("log.pretty") {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: !viper.GetBool("log.color"),
			},
		)
	}

	return nil
}

func setupConfig() {
	if viper.GetString("config.file") != "" {
		viper.SetConfigFile(viper.GetString("config.file"))
	} else {
		viper.SetConfigName("api")
		viper.AddConfigPath("/etc/gexec")
		viper.AddConfigPath("$HOME/.gexec")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("gexec")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := readConfig(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to read config file")
	}

	if err := viper.Unmarshal(cfg); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to parse config file")
	}

	{
		passphrase, err := config.Value(cfg.Encrypt.Passphrase)

		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Failed to parse encrypt passphrase secret")

			os.Exit(1)
		}

		if len(passphrase) != 32 {
			log.Fatal().
				Msg("Encryption passphrase got to be 32 chars")

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
