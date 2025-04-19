package command

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/gexec/gexec/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	dbCmd = &cobra.Command{
		Use:   "database",
		Short: "Database migrations",
		Args:  cobra.NoArgs,
	}

	dbCleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Cleanup expired content",
		Run:   dbCleanupAction,
		Args:  cobra.NoArgs,
	}

	dbMigrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Execute migrations",
		Run:   dbMigrateAction,
		Args:  cobra.NoArgs,
	}

	dbRollbackCmd = &cobra.Command{
		Use:   "rollback",
		Short: "Rollback migrations",
		Run:   dbRollbackAction,
		Args:  cobra.NoArgs,
	}

	dbLockCmd = &cobra.Command{
		Use:   "lock",
		Short: "Lock migrations",
		Run:   dbLockAction,
		Args:  cobra.NoArgs,
	}

	dbUnlockCmd = &cobra.Command{
		Use:   "unlock",
		Short: "Unlock migrations",
		Run:   dbUnlockAction,
		Args:  cobra.NoArgs,
	}

	dbStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Status of migrations",
		Run:   dbStatusAction,
		Args:  cobra.NoArgs,
	}

	dbCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create new migration",
		Run:   dbCreateAction,
	}
)

func init() {
	rootCmd.AddCommand(dbCmd)

	dbCmd.AddCommand(dbCleanupCmd)
	dbCmd.AddCommand(dbMigrateCmd)
	dbCmd.AddCommand(dbRollbackCmd)
	dbCmd.AddCommand(dbLockCmd)
	dbCmd.AddCommand(dbUnlockCmd)
	dbCmd.AddCommand(dbStatusCmd)
	dbCmd.AddCommand(dbCreateCmd)

	dbCmd.PersistentFlags().String("database-driver", defaultDatabaseDriver, "Driver for the database")
	viper.SetDefault("database.driver", defaultDatabaseDriver)
	_ = viper.BindPFlag("database.driver", serverCmd.PersistentFlags().Lookup("database-driver"))

	dbCmd.PersistentFlags().String("database-address", defaultDatabaseAddress, "Address for the database")
	viper.SetDefault("database.address", defaultDatabaseAddress)
	_ = viper.BindPFlag("database.address", serverCmd.PersistentFlags().Lookup("database-address"))

	dbCmd.PersistentFlags().String("database-port", defaultDatabasePort, "Port for the database")
	viper.SetDefault("database.port", defaultDatabasePort)
	_ = viper.BindPFlag("database.port", serverCmd.PersistentFlags().Lookup("database-port"))

	dbCmd.PersistentFlags().String("database-username", defaultDatabaseUsername, "Username for the database")
	viper.SetDefault("database.username", defaultDatabaseUsername)
	_ = viper.BindPFlag("database.username", serverCmd.PersistentFlags().Lookup("database-username"))

	dbCmd.PersistentFlags().String("database-password", defaultDatabasePassword, "Password for the database")
	viper.SetDefault("database.password", defaultDatabasePassword)
	_ = viper.BindPFlag("database.password", serverCmd.PersistentFlags().Lookup("database-password"))

	dbCmd.PersistentFlags().String("database-name", defaultDatabaseName, "Name of the database or path for local databases")
	viper.SetDefault("database.name", defaultDatabaseName)
	_ = viper.BindPFlag("database.name", serverCmd.PersistentFlags().Lookup("database-name"))

	dbCmd.PersistentFlags().StringToString("database-options", defaultDatabaseOptions, "Options for the database connection")
	viper.SetDefault("database.options", defaultDatabaseOptions)
	_ = viper.BindPFlag("database.options", serverCmd.PersistentFlags().Lookup("database-options"))
}

func dbCleanupAction(ccmd *cobra.Command, _ []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	if err := storage.Users.CleanupRedirectTokens(
		context.Background(),
	); err != nil {
		slog.Error(
			"Failed to cleanup redirect tokens",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	slog.Info(
		"Finished cleanup task",
	)
}

func dbMigrateAction(ccmd *cobra.Command, _ []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	group, err := storage.Migrate(ccmd.Context())

	if err != nil {
		slog.Error(
			"Failed to migrate database",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	if group.IsZero() {
		slog.Info(
			"Nothing to migrate",
		)
	} else {
		slog.Info(
			"Finished migrate",
		)
	}
}

func dbRollbackAction(ccmd *cobra.Command, _ []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	group, err := storage.Rollback(ccmd.Context())

	if err != nil {
		slog.Error(
			"Failed to rollback database",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	if group.IsZero() {
		slog.Info(
			"Nothing to rollback",
		)
	} else {
		slog.Info(
			"Finished rollback",
		)
	}
}

func dbLockAction(ccmd *cobra.Command, _ []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	migrator, err := storage.Migrator(
		ccmd.Context(),
	)

	if err != nil {
		slog.Error(
			"Failed to init migrator",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	if err := migrator.Lock(ccmd.Context()); err != nil {
		slog.Error(
			"Failed to lock database",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	slog.Debug(
		"Finished locking",
	)
}

func dbUnlockAction(ccmd *cobra.Command, _ []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	migrator, err := storage.Migrator(
		ccmd.Context(),
	)

	if err != nil {
		slog.Error(
			"Failed to init migrator",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	if err := migrator.Unlock(ccmd.Context()); err != nil {
		slog.Error(
			"Failed to unlock database",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	slog.Debug(
		"Finished unlocking",
	)
}

func dbStatusAction(ccmd *cobra.Command, _ []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	migrator, err := storage.Migrator(
		ccmd.Context(),
	)

	if err != nil {
		slog.Error(
			"Failed to init migrator",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	m, err := migrator.MigrationsWithStatus(
		ccmd.Context(),
	)

	if err != nil {
		slog.Error(
			"Failed to check migrations",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	pending := []string{}

	for _, row := range m.Unapplied() {
		pending = append(pending, row.String())
	}

	applied := []string{}

	for _, row := range m.Applied() {
		applied = append(applied, row.String())
	}

	slog.Info(
		"Migrations",
		slog.Any("pending", pending),
		slog.Any("applied", applied),
	)
}

func dbCreateAction(ccmd *cobra.Command, args []string) {
	storage := prepareStorage(ccmd.Context())
	defer func() { _, _ = storage.Close() }()

	migrator, err := storage.Migrator(
		ccmd.Context(),
	)

	if err != nil {
		slog.Error(
			"Failed to init migrator",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	m, err := migrator.CreateGoMigration(
		ccmd.Context(),
		strings.Join(
			args,
			"_",
		),
	)

	if err != nil {
		slog.Error(
			"Failed to generate migration",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	slog.Info(
		"Finished generating",
		slog.String("name", m.Name),
		slog.String("path", m.Path),
	)
}

func prepareStorage(ctx context.Context) *store.Store {
	storage, err := store.NewStore(
		cfg.Database,
		cfg.Scim,
		cfg.Encrypt,
		nil,
	)

	if err != nil {
		slog.Error(
			"Failed to setup database",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	slog.Info(
		"Preparing database",
		storage.Info()...,
	)

	if val, err := backoff.Retry(
		ctx,
		storage.Open,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithNotify(func(err error, dur time.Duration) {
			slog.Warn(
				"Database open failed",
				slog.Any("error", err),
				slog.Duration("retry", dur),
			)
		}),
	); err != nil || !val {
		slog.Error(
			"Giving up to connect to db",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	if val, err := backoff.Retry(
		ctx,
		storage.Ping,
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithNotify(func(err error, dur time.Duration) {
			slog.Warn(
				"Database ping failed",
				slog.Any("error", err),
				slog.Duration("retry", dur),
			)
		}),
	); err != nil || !val {
		slog.Error(
			"Giving up to ping the db",
			slog.Any("error", err),
		)

		os.Exit(1)
	}

	return storage
}
