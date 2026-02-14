package config

import (
	"time"

	"github.com/spf13/viper"
)

func ENV() string {
	return viper.GetString("env")
}

func GetDbPort() string {
	return viper.GetString("port")
}

func GetDbHost() string {
	return viper.GetString("postgres.dbhost")
}

func GetDbName() string {
	return viper.GetString("postgres.dbname")
}

func GetDbUser() string {
	return viper.GetString("postgres.dbuser")
}

func GetDbPassword() string {
	return viper.GetString("postgres.dbpass")
}

func JWTSigningKey() string {
	return viper.GetString("jwt.signing_key")
}

func JWTExp() time.Duration {
	return viper.GetDuration("jwt.exp")
}

func GetString(key string) string {
	return viper.GetString(key)
}
