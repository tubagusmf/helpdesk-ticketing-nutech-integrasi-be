package db

import (
	"log"
	"os"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewPostgres() *gorm.DB {
	dsn := helper.GetConnectionString()

	gormLogger := logger.New(
		log.New(
			os.Stdout,
			"\r\n",
			log.LstdFlags,
		),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Error,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{
			Logger: gormLogger,
		},
	)

	if err != nil {
		log.Fatalf(
			"Failed to connect to database: %v",
			err,
		)
	}

	log.Println("PostgreSQL database connected successfully")

	return db
}
