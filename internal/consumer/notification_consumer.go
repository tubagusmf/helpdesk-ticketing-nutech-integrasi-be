package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

func StartNotificationConsumer(
	notificationUsecase model.INotificationUsecase,
) {

	q, err := config.RabbitChannel.QueueDeclare(
		"notification_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	err = config.RabbitChannel.QueueBind(
		q.Name,
		"ticket.*",
		"ticket_events",
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	msgs, err := config.RabbitChannel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for d := range msgs {

			var event model.NotificationEvent

			err := json.Unmarshal(d.Body, &event)
			if err != nil {
				log.Println(err)
				continue
			}

			_, err = notificationUsecase.Create(
				context.Background(),
				model.CreateNotificationInput{
					UserID:        event.UserID,
					ActorID:       event.ActorID,
					TicketID:      event.TicketID,
					Type:          model.NotificationType(event.EventType),
					ReferenceType: model.NotificationReferenceType(event.ReferenceType),
					ReferenceID:   event.ReferenceID,
					Title:         event.Title,
					Message:       event.Message,
				},
			)

			if err != nil {
				log.Println("failed create notification:", err)
			}
		}
	}()
}
