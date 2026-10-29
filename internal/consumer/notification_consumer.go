package consumer

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"firebase.google.com/go/v4/messaging"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/fcm"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

func StartNotificationConsumer(notificationUsecase model.INotificationUsecase, fcmTokenUsecase model.IUserFCMTokenUsecase, fcmService *fcm.Service) {
	if config.RabbitChannel == nil {
		log.Println("[NOTIFICATION] RabbitMQ channel belum tersedia")
		return
	}

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

	if fcmService == nil {
		log.Println(
			"[NOTIFICATION] FCM tidak aktif; " +
				"notifikasi tetap disimpan ke database",
		)
	}

	go func() {
		for d := range msgs {
			var event model.NotificationEvent

			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf(
					"[NOTIFICATION] Payload tidak valid: %v",
					err,
				)
				continue
			}

			ctx, cancel := context.WithTimeout(
				context.Background(),
				10*time.Second,
			)

			notification, err := notificationUsecase.Create(
				ctx,
				model.CreateNotificationInput{
					UserID:   event.UserID,
					ActorID:  event.ActorID,
					TicketID: event.TicketID,
					Type: model.NotificationType(
						event.EventType,
					),
					ReferenceType: model.NotificationReferenceType(
						event.ReferenceType,
					),
					ReferenceID: event.ReferenceID,
					Title:       event.Title,
					Message:     event.Message,
				},
			)

			cancel()

			if err != nil {
				log.Printf(
					"[NOTIFICATION] Gagal menyimpan notification: %v",
					err,
				)
				continue
			}

			if fcmService == nil || fcmTokenUsecase == nil {
				continue
			}

			if event.UserID <= 0 {
				log.Printf(
					"[FCM] UserID tidak valid untuk ticket %d",
					event.TicketID,
				)
				continue
			}

			lookupCtx, lookupCancel := context.WithTimeout(
				context.Background(),
				5*time.Second,
			)

			devices, err := fcmTokenUsecase.FindActiveByUserID(
				lookupCtx,
				event.UserID,
			)

			lookupCancel()

			if err != nil {
				log.Printf(
					"[FCM] Gagal mengambil perangkat user %d: %v",
					event.UserID,
					err,
				)
				continue
			}

			if len(devices) == 0 {
				log.Printf(
					"[FCM] Tidak ada perangkat terdaftar untuk user %d",
					event.UserID,
				)
				continue
			}

			data := map[string]string{
				"type": event.EventType,
				"notification_id": strconv.FormatInt(
					notification.ID, 10,
				),
				"ticket_code": event.TicketCode,
			}

			if event.TicketID > 0 {
				data["ticket_id"] = strconv.FormatInt(
					event.TicketID, 10,
				)
			}

			for _, device := range devices {
				sendCtx, sendCancel := context.WithTimeout(
					context.Background(),
					10*time.Second,
				)

				sendErr := fcmService.SendToToken(
					sendCtx,
					device.Token,
					event.Title,
					event.Message,
					data,
				)

				sendCancel()

				if sendErr != nil {
					log.Printf(
						"[FCM] Gagal mengirim push notification "+
							"notification_id=%d user_id=%d: %v",
						notification.ID,
						event.UserID,
						sendErr,
					)

					if messaging.IsUnregistered(sendErr) {
						cleanupCtx, cleanupCancel :=
							context.WithTimeout(
								context.Background(),
								5*time.Second,
							)

						deleteErr :=
							fcmTokenUsecase.DeleteInvalidToken(
								cleanupCtx,
								device.Token,
							)

						cleanupCancel()

						if deleteErr != nil {
							log.Printf(
								"[FCM] Gagal menghapus token invalid: %v",
								deleteErr,
							)
						}
					}

					continue
				}

				log.Printf(
					"[FCM] Push berhasil dikirim "+
						"notification_id=%d user_id=%d",
					notification.ID,
					event.UserID,
				)
			}
		}
	}()
}
