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

	ticket, err := u.ticketRepo.FindByID(ctx, comment.TicketID)
	if err != nil {
		log.Error("Failed find ticket:", err)
		return nil, err
	}

	err = helper.PublishNotificationEvent(
		"ticket.comment",
		model.NotificationEvent{
			EventType:     "TICKET_COMMENT",
			UserID:        ticket.ReporterID,
			ActorID:       comment.UserID,
			TicketID:      ticket.ID,
			TicketCode:    ticket.TicketCode,
			ReferenceType: "COMMENT",
			ReferenceID:   result.ID,
			Title:         "New Comment",
			Message:       "New comment added to ticket " + ticket.TicketCode,
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
