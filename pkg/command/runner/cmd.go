package command

import (
	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:           "gexec-runner",
		Short:         "Generic execution platform for Ansible and Terraform/OpenTofu",
		Version:       version.String,
		SilenceErrors: false,
		SilenceUsage:  true,

		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return setupLogger()
		},

		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cfg *config.Config
)

func init() {
	cfg = config.Load()
	cobra.OnInitialize(setupConfig)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Show the help, so what you see now")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the current version of that tool")

	rootCmd.PersistentFlags().String("config-file", "", "Path to optional config file")
	_ = viper.BindPFlag("config.file", rootCmd.PersistentFlags().Lookup("config-file"))

	rootCmd.PersistentFlags().String("log-level", "info", "Set logging level")
	viper.SetDefault("log.level", "info")
	_ = viper.BindPFlag("log.level", rootCmd.PersistentFlags().Lookup("log-level"))

	rootCmd.PersistentFlags().Bool("log-pretty", true, "Enable pretty logging")
	viper.SetDefault("log.pretty", true)
	_ = viper.BindPFlag("log.pretty", rootCmd.PersistentFlags().Lookup("log-pretty"))
}

// Run parses the command line arguments and executes the program.
func Run() error {
	return rootCmd.Execute()
}
