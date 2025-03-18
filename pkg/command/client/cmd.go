package command

import (
	"html/template"

	"github.com/drone/funcmap"
	"github.com/gexec/gexec/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultServerAddress = "http://localhost:8080/api/v1"
)

var (
	rootCmd = &cobra.Command{
		Use:           "gexec-client",
		Short:         "Generic execution platform for Ansible/OpenTofu/Terraform",
		Version:       version.String,
		SilenceErrors: false,
		SilenceUsage:  true,

		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	// basicFuncMap provides template helpers provided by library.
	basicFuncMap = funcmap.Funcs

	// globalFuncMap provides global template helper functions.
	globalFuncMap = template.FuncMap{}
)

func init() {
	cobra.OnInitialize(setupConfig)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Show the help, so what you see now")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the current version of that tool")

	rootCmd.PersistentFlags().StringP("server-address", "s", defaultServerAddress, "Server address")
	_ = viper.BindPFlag("server.address", rootCmd.PersistentFlags().Lookup("server-address"))

	rootCmd.PersistentFlags().StringP("server-token", "t", "", "Server token")
	_ = viper.BindPFlag("server.token", rootCmd.PersistentFlags().Lookup("server-token"))

	rootCmd.PersistentFlags().StringP("server-username", "u", "", "Server username")
	_ = viper.BindPFlag("server.username", rootCmd.PersistentFlags().Lookup("server-username"))

	rootCmd.PersistentFlags().StringP("server-password", "p", "", "Server password")
	_ = viper.BindPFlag("server.password", rootCmd.PersistentFlags().Lookup("server-password"))
}

// Run parses the command line arguments and executes the program.
func Run() error {
	return rootCmd.Execute()
}
