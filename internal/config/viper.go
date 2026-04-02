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
		log.Println("Config.yml not found, skipping...")
	}

	viper.AutomaticEnv()
}
