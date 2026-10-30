package usecase

import (
	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

func ticketNotificationRecipients(ticket *model.Ticket, actorID int64) []int64 {
	if ticket == nil {
		return nil
	}

	recipients := make([]int64, 0, 3)
	seen := make(map[int64]struct{})

	addRecipient := func(id int64) {
		if id <= 0 || id == actorID {
			return
		}

		if _, exists := seen[id]; exists {
			return
		}

		seen[id] = struct{}{}
		recipients = append(recipients, id)
	}

	addRecipient(ticket.ReporterID)

	if ticket.StaffAssignedToID != nil {
		addRecipient(*ticket.StaffAssignedToID)
	}

	if ticket.AssignedToID != nil {
		addRecipient(*ticket.AssignedToID)
	}

	return recipients
}

func publishTicketNotification(routingKey string, ticket *model.Ticket, actorID int64, notificationType model.NotificationType, referenceType model.NotificationReferenceType, referenceID int64, title string, message string) {
	if ticket == nil {
		return
	}

	recipients := ticketNotificationRecipients(
		ticket,
		actorID,
	)

	for _, recipientID := range recipients {
		err := helper.PublishNotificationEvent(
			routingKey,
			model.NotificationEvent{
				EventType: string(notificationType),

				UserID:  recipientID,
				ActorID: actorID,

				TicketID:   ticket.ID,
				TicketCode: ticket.TicketCode,

				ReferenceType: string(referenceType),
				ReferenceID:   referenceID,

				Title:   title,
				Message: message,
			},
		)

		if err != nil {
			logrus.WithFields(logrus.Fields{
				"ticket_id":    ticket.ID,
				"recipient_id": recipientID,
				"actor_id":     actorID,
				"event":        routingKey,
			}).WithError(err).Error(
				"Gagal mengirim notifikasi tiket",
			)
		}
	}
}

func publishTicketStatusNotification(ticket *model.Ticket, actorID int64, oldStatus model.TicketStatus, newStatus model.TicketStatus) {
	if ticket == nil || oldStatus == newStatus {
		return
	}

	notificationType := model.NotificationTicketUpdated
	title := "Status Tiket Diperbarui"

	switch newStatus {
	case model.StatusResolved:
		notificationType = model.NotificationTicketResolved
		title = "Tiket Diselesaikan"

	case model.StatusClosed:
		notificationType = model.NotificationTicketClosed
		title = "Tiket Ditutup"
	}

	message := "No Tiket: " +
		ticket.TicketCode +
		" | Status tiket berubah dari " +
		string(oldStatus) +
		" menjadi " +
		string(newStatus)

	publishTicketNotification(
		"ticket.status",
		ticket,
		actorID,
		notificationType,
		model.ReferenceTicket,
		ticket.ID,
		title,
		message,
	)
}
