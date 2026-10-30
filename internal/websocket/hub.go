package websocket

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

type PresenceUpdater interface {
	UpdateOnlineStatus(ctx context.Context, userID int64, isOnline bool) error
	MarkStaleUsersOffline(ctx context.Context, cutoff time.Time) error
}

type RoleMessage struct {
	Role    string
	Message []byte
}

type presenceUpdate struct {
	UserID   int64
	IsOnline bool
}

type Hub struct {
	Clients         map[string]map[*Client]bool
	UserClients     map[int64]map[*Client]bool
	Register        chan *Client
	Unregister      chan *Client
	BroadcastToRole chan RoleMessage
	presenceUpdater PresenceUpdater
	presenceUpdates chan presenceUpdate
}

func NewHub(updaters ...PresenceUpdater) *Hub {
	var updater PresenceUpdater

	if len(updaters) > 0 {
		updater = updaters[0]
	}

	return &Hub{
		Clients: map[string]map[*Client]bool{
			"ADMINISTRATOR": {},
			"STAFF":         {},
			"USER":          {},
			"EXECUTIVE":     {},
			"ENGINEER":      {},
		},

		UserClients: make(map[int64]map[*Client]bool),

		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		BroadcastToRole: make(chan RoleMessage),

		presenceUpdater: updater,
		presenceUpdates: make(chan presenceUpdate, 256),
	}
}

func (h *Hub) Run() {
	h.resetOnlineUsersOnStartup()

	go h.runPresenceUpdates()

	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)

		case client := <-h.Unregister:
			h.unregisterClient(client)

		case msg := <-h.BroadcastToRole:
			clients := h.Clients[msg.Role]

			logrus.Infof(
				"[WS BROADCAST ROLE] role=%s clients=%d",
				msg.Role,
				len(clients),
			)

			for client := range clients {
				select {
				case client.Send <- msg.Message:

				default:
					logrus.Warn(
						"[WS] Client channel full, removing client",
					)

					h.unregisterClient(client)

					if client.Conn != nil {
						_ = client.Conn.Close()
					}
				}
			}
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	if client == nil || client.UserID <= 0 {
		return
	}

	if h.Clients[client.Role] == nil {
		h.Clients[client.Role] = make(map[*Client]bool)
	}

	if h.UserClients[client.UserID] == nil {
		h.UserClients[client.UserID] = make(map[*Client]bool)
	}

	if h.UserClients[client.UserID][client] {
		return
	}

	firstConnection := len(h.UserClients[client.UserID]) == 0

	h.Clients[client.Role][client] = true
	h.UserClients[client.UserID][client] = true

	logrus.Infof(
		"[WS REGISTER] user_id=%d role=%s connections=%d",
		client.UserID,
		client.Role,
		len(h.UserClients[client.UserID]),
	)

	if firstConnection {
		h.queuePresence(client.UserID, true)
	}
}

func (h *Hub) unregisterClient(client *Client) {
	if client == nil {
		return
	}

	roleClients, roleExists := h.Clients[client.Role]

	if !roleExists {
		return
	}

	if _, exists := roleClients[client]; !exists {
		return
	}

	delete(roleClients, client)

	close(client.Send)

	userClients, exists := h.UserClients[client.UserID]

	if !exists {
		return
	}

	delete(userClients, client)

	logrus.Infof(
		"[WS UNREGISTER] user_id=%d remaining_connections=%d",
		client.UserID,
		len(userClients),
	)

	if len(userClients) == 0 {
		delete(h.UserClients, client.UserID)

		h.queuePresence(client.UserID, false)
	}
}

func (h *Hub) queuePresence(userID int64, isOnline bool) {
	if h.presenceUpdater == nil || userID <= 0 {
		return
	}

	h.presenceUpdates <- presenceUpdate{
		UserID:   userID,
		IsOnline: isOnline,
	}
}

func (h *Hub) runPresenceUpdates() {
	if h.presenceUpdater == nil {
		return
	}

	for update := range h.presenceUpdates {
		ctx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)

		err := h.presenceUpdater.UpdateOnlineStatus(
			ctx,
			update.UserID,
			update.IsOnline,
		)

		cancel()

		if err != nil {
			logrus.Errorf(
				"[WS PRESENCE] failed update user=%d online=%t: %v",
				update.UserID,
				update.IsOnline,
				err,
			)
			continue
		}

		logrus.Infof(
			"[WS PRESENCE] user_id=%d is_online=%t",
			update.UserID,
			update.IsOnline,
		)
	}
}

func (h *Hub) resetOnlineUsersOnStartup() {
	if h.presenceUpdater == nil {
		return
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	cutoff := time.Now().Add(time.Second)

	if err := h.presenceUpdater.MarkStaleUsersOffline(
		ctx,
		cutoff,
	); err != nil {
		logrus.Errorf(
			"[WS PRESENCE] failed reset online users on startup: %v",
			err,
		)
		return
	}

	logrus.Info(
		"[WS PRESENCE] previous online sessions reset successfully",
	)
}
