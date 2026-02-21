package console

import (
	"log"
	"net/http"
	"sync"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/db"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/repository"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	handlerHttp "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/delivery/http"
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

	userRepo := repository.NewUserRepo(postgresDB)
	roleRepo := repository.NewRoleRepo(postgresDB)
	projectRepo := repository.NewProjectRepo(postgresDB)
	locationRepo := repository.NewLocationRepo(postgresDB)
	partRepo := repository.NewPartRepo(postgresDB)
	assetIDRepo := repository.NewAssetIDRepo(postgresDB)
	causeRepo := repository.NewCauseRepo(postgresDB)
	solutionRepo := repository.NewSolutionRepo(postgresDB)
	ticketRepo := repository.NewTicketRepo(postgresDB)
	ticketHistoryRepo := repository.NewTicketHistoryRepo(postgresDB)
	ticketComment := repository.NewTicketCommentRepo(postgresDB)
	ticketResolution := repository.NewTicketResolutionRepo(postgresDB)

	userUsecase := usecase.NewUserUsecase(userRepo)
	roleUsecase := usecase.NewRoleUsecase(roleRepo)
	projectUsecase := usecase.NewProjectUsecase(projectRepo)
	locationUsecase := usecase.NewLocationUsecase(locationRepo)
	partUsecase := usecase.NewPartUsecase(partRepo)
	assetIDUsecase := usecase.NewAssetIDUsecase(assetIDRepo)
	causeUsecase := usecase.NewCauseUsecase(causeRepo)
	solutionUsecase := usecase.NewSolutionUsecase(solutionRepo)
	ticketUsecase := usecase.NewTicketUsecase(postgresDB, ticketRepo, ticketHistoryRepo)
	ticketHistoryUsecase := usecase.NewTicketHistoryUsecase(ticketHistoryRepo)
	ticketCommentUsecase := usecase.NewTicketCommentUsecase(ticketComment)
	ticketResolutionUsecase := usecase.NewTicketResolutionUsecase(postgresDB, ticketResolution, ticketHistoryRepo, ticketRepo)

	e := echo.New()

	handlerHttp.NewUserHandler(e, userUsecase)
	handlerHttp.NewRoleHandler(e, roleUsecase)
	handlerHttp.NewProjectHandler(e, projectUsecase)
	handlerHttp.NewLocationHandler(e, locationUsecase)
	handlerHttp.NewPartHandler(e, partUsecase)
	handlerHttp.NewAssetIDHandler(e, assetIDUsecase)
	handlerHttp.NewCauseHandler(e, causeUsecase)
	handlerHttp.NewSolutionHandler(e, solutionUsecase)
	handlerHttp.NewTicketHandler(e, ticketUsecase)
	handlerHttp.NewTicketHistoryHandler(e, ticketHistoryUsecase)
	handlerHttp.NewTicketCommentHandler(e, ticketCommentUsecase)
	handlerHttp.NewTicketResolutionHandler(e, ticketResolutionUsecase)

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
