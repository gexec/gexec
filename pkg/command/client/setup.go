package command

import (
	"strings"

	"github.com/spf13/viper"
)

func setupConfig() {
	viper.SetEnvPrefix("gexec")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
