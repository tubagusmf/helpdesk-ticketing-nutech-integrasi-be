package helper

import (
	"encoding/json"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishNotificationEvent(
	routingKey string,
	event model.NotificationEvent,
) error {

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return config.RabbitChannel.Publish(
		"ticket_events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
