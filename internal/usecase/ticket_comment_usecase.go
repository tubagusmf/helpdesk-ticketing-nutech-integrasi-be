package usecase

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketCommentUsecase struct {
	repo model.ITicketCommentRepository
}

func NewTicketCommentUsecase(
	repo model.ITicketCommentRepository,
) model.ITicketCommentUsecase {
	return &TicketCommentUsecase{
		repo: repo,
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

	return result, nil
}

func (u *TicketCommentUsecase) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketComment, error) {
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
