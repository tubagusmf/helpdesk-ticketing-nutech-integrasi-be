package usecase

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketCommentUsecase struct {
	repo              model.ITicketCommentRepository
	ticketHistoryRepo model.ITicketHistoryRepository
	ticketRepo        model.ITicketRepository
}

func NewTicketCommentUsecase(
	repo model.ITicketCommentRepository,
	ticketHistoryRepo model.ITicketHistoryRepository,
	ticketRepo model.ITicketRepository,
) model.ITicketCommentUsecase {
	return &TicketCommentUsecase{
		repo:              repo,
		ticketHistoryRepo: ticketHistoryRepo,
		ticketRepo:        ticketRepo,
	}
}

func (u *TicketCommentUsecase) Create(ctx context.Context, comment model.TicketComment) (*model.TicketComment, error) {
	log := logrus.WithFields(logrus.Fields{
		"comment": comment,
	})

	ticket, err := u.ticketRepo.FindByID(ctx, comment.TicketID)
	if err != nil {
		log.Error("Failed find ticket:", err)
		return nil, err
	}

	if comment.UserID == ticket.ReporterID {
		comment.IsReadByUser = true
		comment.IsReadByStaff = false
		comment.IsReadByAdministrator = false
	} else {
		comment.IsReadByUser = false
		comment.IsReadByStaff = true
		comment.IsReadByAdministrator = true
	}

	result, err := u.repo.Create(ctx, comment)
	if err != nil {
		log.Error("Failed to create ticket comment: ", err)
		return nil, err
	}

	message := comment.Message

	_, err = u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
		TicketID: comment.TicketID,
		UserID:   comment.UserID,
		Action:   "COMMENT",
		NewValue: &message,
	})
	if err != nil {
		log.Error("FAILED insert history COMMENT: ", err)
	} else {
		log.Info("SUCCESS insert history COMMENT")
	}

	var receiverID int64

	if comment.UserID == ticket.ReporterID {

		if ticket.AssignedToID == nil {
			log.Warn("Ticket has no assigned staff")
			return result, nil
		}

		receiverID = *ticket.AssignedToID

	} else {
		receiverID = ticket.ReporterID
	}

	messagePreview := comment.Message

	if len(messagePreview) > 80 {
		messagePreview = messagePreview[:80] + "..."
	}

	err = helper.PublishNotificationEvent(
		"ticket.comment",
		model.NotificationEvent{
			EventType:     "TICKET_COMMENT",
			UserID:        receiverID,
			ActorID:       comment.UserID,
			TicketID:      ticket.ID,
			TicketCode:    ticket.TicketCode,
			ReferenceType: "COMMENT",
			ReferenceID:   result.ID,
			Title:         "Komentar Baru",
			Message:       "Komentar pada tiket " + ticket.TicketCode + ": " + messagePreview,
		},
	)

	if err != nil {
		log.Error("Failed publish notification:", err)
	}

	return result, nil
}

func (u *TicketCommentUsecase) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketCommentResponse, error) {
	log := logrus.WithFields(logrus.Fields{
		"ticketID": ticketID,
	})

	comment, err := u.repo.FindByTicketID(ctx, ticketID)
	if err != nil {
		log.Error("Failed to find ticket comment: ", err)
		return nil, err
	}

	return comment, nil
}

func (u *TicketCommentUsecase) MarkAsRead(ctx context.Context, ticketID int64, role string) error {
	return u.repo.MarkAsRead(ctx, ticketID, role)
}
