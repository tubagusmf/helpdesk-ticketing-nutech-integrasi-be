package websocket

import "github.com/sirupsen/logrus"

type Hub struct {
	Clients map[string]map[*Client]bool

	Register   chan *Client
	Unregister chan *Client

	BroadcastToRole chan RoleMessage
}

type RoleMessage struct {
	Role    string
	Message []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients: map[string]map[*Client]bool{
			"ADMINISTRATOR": {},
			"STAFF":         {},
			"USER":          {},
		},

		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		BroadcastToRole: make(chan RoleMessage),
	}
}

func (h *Hub) Run() {
	for {

		select {

		case client := <-h.Register:

			if _, ok := h.Clients[client.Role]; !ok {
				h.Clients[client.Role] = make(map[*Client]bool)
			}

			h.Clients[client.Role][client] = true

		case client := <-h.Unregister:

			if clients, ok := h.Clients[client.Role]; ok {

				if _, exists := clients[client]; exists {

					delete(clients, client)

					close(client.Send)
				}
			}

		case msg := <-h.BroadcastToRole:

			if clients, ok := h.Clients[msg.Role]; ok {

				for client := range clients {

					select {

					case client.Send <- msg.Message:

					default:

						logrus.Warn(
							"CLIENT CHANNEL FULL, REMOVED",
						)

						close(client.Send)
						delete(clients, client)
					}
				}
			}
		}
	}
}
