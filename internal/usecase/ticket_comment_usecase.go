package usecase

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketCommentUsecase struct {
	repo              model.ITicketCommentRepository
	ticketHistoryRepo model.ITicketHistoryRepository
}

func NewTicketCommentUsecase(
	repo model.ITicketCommentRepository,
	ticketHistoryRepo model.ITicketHistoryRepository,
) model.ITicketCommentUsecase {
	return &TicketCommentUsecase{
		repo:              repo,
		ticketHistoryRepo: ticketHistoryRepo,
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
