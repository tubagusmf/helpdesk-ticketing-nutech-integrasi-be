package config

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func getEnvOrViper(envKey, configKey string) string {
	value := strings.TrimSpace(os.Getenv(envKey))

	if value != "" {
		return value
	}

	return viper.GetString(configKey)
}

func ENV() string {
	return getEnvOrViper("APP_ENV", "env")
}

func GetDbPort() string {
	return getEnvOrViper("POSTGRES_PORT", "port")
}

func GetDbHost() string {
	return getEnvOrViper("POSTGRES_HOST", "postgres.dbhost")
}

func GetDbName() string {
	return getEnvOrViper("POSTGRES_DBNAME", "postgres.dbname")
}

func GetDbUser() string {
	return getEnvOrViper("POSTGRES_DBUSER", "postgres.dbuser")
}

func GetDbPassword() string {
	return getEnvOrViper("POSTGRES_DBPASS", "postgres.dbpass")
}

func JWTSigningKey() string {
	return getEnvOrViper(
		"JWT_SIGNING_KEY",
		"jwt.signing_key",
	)
}

func JWTExp() time.Duration {
	value := strings.TrimSpace(os.Getenv("JWT_EXP"))

	if value != "" {
		duration, err := time.ParseDuration(value)
		if err == nil {
			return duration
		}
	}

	return viper.GetDuration("jwt.exp")
}

func GetString(key string) string {
	return viper.GetString(key)
}
