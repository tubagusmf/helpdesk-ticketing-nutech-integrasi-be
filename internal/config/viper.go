package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadWithViper() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config.yml: %v", err)
	}

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.MergeInConfig(); err != nil {
		log.Fatalf("Error reading .env: %v", err)
	}

	viper.AutomaticEnv()
}
