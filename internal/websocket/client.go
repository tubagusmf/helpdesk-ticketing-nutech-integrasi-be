package websocket

import (
	"log"
	"time"

	gws "github.com/gorilla/websocket"
)

type Client struct {
	Conn *gws.Conn
	Send chan []byte

	UserID int64
	Role   string
}

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)

	c.Conn.SetReadDeadline(
		time.Now().Add(60 * time.Second),
	)

	c.Conn.SetPongHandler(func(string) error {

		c.Conn.SetReadDeadline(
			time.Now().Add(60 * time.Second),
		)

		return nil
	})

	for {

		_, _, err := c.Conn.ReadMessage()

		if err != nil {

			if gws.IsUnexpectedCloseError(
				err,
				gws.CloseGoingAway,
				gws.CloseAbnormalClosure,
			) {
				log.Println("websocket error:", err)
			}

			break
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {

		select {

		case message, ok := <-c.Send:

			c.Conn.SetWriteDeadline(
				time.Now().Add(10 * time.Second),
			)

			if !ok {

				c.Conn.WriteMessage(
					gws.CloseMessage,
					[]byte{},
				)

				return
			}

			err := c.Conn.WriteMessage(
				gws.TextMessage,
				message,
			)

			if err != nil {
				return
			}

		case <-ticker.C:

			c.Conn.SetWriteDeadline(
				time.Now().Add(10 * time.Second),
			)

			if err := c.Conn.WriteMessage(
				gws.PingMessage,
				nil,
			); err != nil {

				return
			}
		}
	}
}
