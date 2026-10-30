package websocket

import (
	"log"
	"time"

	gws "github.com/gorilla/websocket"
)

const (
	pongWait   = 50 * time.Second
	pingPeriod = 20 * time.Second
	writeWait  = 10 * time.Second
	maxMessage = 512
)

type Client struct {
	Conn      *gws.Conn
	Send      chan []byte
	UserID    int64
	Role      string
	ExpiresAt time.Time
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessage)

	c.Conn.SetReadDeadline(
		time.Now().Add(pongWait),
	)

	c.Conn.SetPongHandler(func(string) error {
		return c.Conn.SetReadDeadline(
			time.Now().Add(pongWait),
		)
	})

	for {
		_, _, err := c.Conn.ReadMessage()

		if err != nil {
			if gws.IsUnexpectedCloseError(
				err,
				gws.CloseGoingAway,
				gws.CloseAbnormalClosure,
			) {
				log.Println(
					"[WS] read error:",
					err,
				)
			}

			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	var expiryTimer *time.Timer
	var expiryChannel <-chan time.Time

	if !c.ExpiresAt.IsZero() {
		remaining := time.Until(c.ExpiresAt)

		if remaining <= 0 {
			return
		}

		expiryTimer = time.NewTimer(remaining)
		expiryChannel = expiryTimer.C

		defer expiryTimer.Stop()
	}

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(
				time.Now().Add(writeWait),
			)

			if !ok {
				_ = c.Conn.WriteControl(
					gws.CloseMessage,
					gws.FormatCloseMessage(
						gws.CloseNormalClosure,
						"connection closed",
					),
					time.Now().Add(writeWait),
				)

				return
			}

			if err := c.Conn.WriteMessage(
				gws.TextMessage,
				message,
			); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.Conn.WriteControl(
				gws.PingMessage,
				nil,
				time.Now().Add(writeWait),
			); err != nil {
				return
			}

		case <-expiryChannel:
			_ = c.Conn.WriteControl(
				gws.CloseMessage,
				gws.FormatCloseMessage(
					gws.CloseNormalClosure,
					"JWT expired",
				),
				time.Now().Add(writeWait),
			)

			return
		}
	}
}
