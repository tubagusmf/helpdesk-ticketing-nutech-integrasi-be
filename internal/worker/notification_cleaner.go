package worker

import (
	"context"
	"log"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type NotificationCleaner struct {
	notificationUsecase model.INotificationUsecase
}

func NewNotificationCleaner(
	notificationUsecase model.INotificationUsecase,
) *NotificationCleaner {
	return &NotificationCleaner{
		notificationUsecase: notificationUsecase,
	}
}

func (w *NotificationCleaner) Start() {
	ticker := time.NewTicker(1 * time.Minute)

	defer ticker.Stop()

	for range ticker.C {

		expiredBefore := time.Now().Add(-15 * time.Minute)

		err := w.notificationUsecase.DeleteExpired(
			context.Background(),
			expiredBefore,
		)

		if err != nil {
			log.Println("[NOTIF CLEANER ERROR]", err)
			continue
		}
	}
}
