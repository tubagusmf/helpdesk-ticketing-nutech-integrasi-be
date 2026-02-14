package config

import "github.com/sirupsen/logrus"

func SetupLogger() {
	log := logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})
}
