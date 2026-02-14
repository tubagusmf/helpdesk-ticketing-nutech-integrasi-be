package console

import (
	"log"
	"net/http"
	"sync"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/db"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCMD)
}

var serverCMD = &cobra.Command{
	Use:   "httpsrv",
	Short: "Start HTTP server",
	Long:  "Start the HTTP server to handle incoming requests for the to-do list application.",
	Run:   httpServer,
}

func httpServer(cmd *cobra.Command, args []string) {
	config.LoadWithViper()

	postgresDB := db.NewPostgres()
	sqlDB, err := postgresDB.DB()
	if err != nil {
		log.Fatalf("Failed to get SQL DB from Gorm: %v", err)
	}
	defer sqlDB.Close()

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	wg.Add(2)

	go func() {
		defer wg.Done()
		errCh <- e.Start(":3000")
	}()

	go func() {
		defer wg.Done()
		<-errCh
	}()

	wg.Wait()

	if err := <-errCh; err != nil {
		if err != http.ErrServerClosed {
			logrus.Errorf("HTTP server error: %v", err)
		}
	}
}
