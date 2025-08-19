package config

import (
	"strings"

	"github.com/spf13/viper"
)

func bindEnvironment(v *viper.Viper) {
	v.SetEnvPrefix("sfDBTools")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}
