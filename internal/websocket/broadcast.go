package websocket

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

func BroadcastToRoles(
	hub *Hub,
	roles []string,
	message Message,
) {

	payload, err := json.Marshal(message)

	if err != nil {
		logrus.Error(
			"failed marshal websocket message:",
			err,
		)

		return
	}

	for _, role := range roles {

		go func(r string) {

			hub.BroadcastToRole <- RoleMessage{
				Role:    r,
				Message: payload,
			}

		}(role)
	}
}
