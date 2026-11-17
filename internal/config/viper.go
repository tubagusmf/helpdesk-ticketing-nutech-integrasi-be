package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

func LoadWithViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvKeyReplacer(
		strings.NewReplacer(".", "_"),
	)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Config.yml not found, skipping...")
	}
}
