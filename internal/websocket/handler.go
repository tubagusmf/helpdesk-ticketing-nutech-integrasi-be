package websocket

import (
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	Hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		Hub: hub,
	}
}

var upgrader = websocket.Upgrader{
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

	token, err := jwt.Parse(
		tokenString,
		func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, echo.NewHTTPError(
					http.StatusUnauthorized,
					"invalid signing method",
				)
			}

			return []byte(os.Getenv("JWT_SECRET")), nil
		},
	)

	if err != nil || !token.Valid {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid token",
		)
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid claims",
		)
	}

	role, ok := claims["role"].(string)

	if !ok {
		return echo.NewHTTPError(
			http.StatusUnauthorized,
			"invalid role",
		)
	}

	userID := int64(claims["user_id"].(float64))

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

		UserID: userID,
		Role:   role,
	}

	h.Hub.Register <- client

	go client.ReadPump(h.Hub)
	go client.WritePump()

	return nil
}
