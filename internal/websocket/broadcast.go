package websocket

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func BroadcastToRoles(hub *Hub, roles []string, message Message) {
	payload, err := json.Marshal(message)

	if err != nil {
		logrus.WithError(err).Error(
			"Failed to marshal WebSocket message",
		)
		return
	}

	logrus.WithFields(logrus.Fields{
		"event": message.Type,
		"roles": roles,
	}).Debug("WebSocket broadcast dispatched")

	for _, role := range roles {
		go func(r string) {
			hub.BroadcastToRole <- RoleMessage{
				Role:    r,
				Message: payload,
			}
		}(role)
	}
}
