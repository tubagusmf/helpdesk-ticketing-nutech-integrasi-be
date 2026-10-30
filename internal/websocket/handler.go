package websocket

import (
	"net/http"
	"strings"
	"time"

	gws "github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type Handler struct {
	Hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		Hub: hub,
	}
}

var upgrader = gws.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) Handle(c echo.Context) error {
	tokenString := c.QueryParam("token")

	if tokenString == "" {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"missing token",
		)
	}

	var claims model.CustomClaims

	if err := helper.DecodeToken(
		tokenString,
		&claims,
	); err != nil {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid or expired token",
		)
	}

	if claims.UserID <= 0 {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid user ID",
		)
	}

	role := strings.ToUpper(
		strings.TrimSpace(claims.Role),
	)

	if role == "" {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid user role",
		)
	}

	if claims.ExpiresAt == nil ||
		!claims.ExpiresAt.Time.After(time.Now()) {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"token expiration missing or expired",
		)
	}

	conn, err := upgrader.Upgrade(
		c.Response(),
		c.Request(),
		nil,
	)

	if err != nil {
		return err
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),

		UserID: claims.UserID,
		Role:   role,

		ExpiresAt: claims.ExpiresAt.Time,
	}

	h.Hub.Register <- client

	go client.ReadPump(h.Hub)
	go client.WritePump()

	return nil
}
