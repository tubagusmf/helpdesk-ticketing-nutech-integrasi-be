package usecase

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketHistoryUsecase struct {
	repo model.ITicketHistoryRepository
}

func NewTicketHistoryUsecase(repo model.ITicketHistoryRepository) model.ITicketHistoryUsecase {
	return &TicketHistoryUsecase{
		repo: repo,
	}
}

func (u *TicketHistoryUsecase) Create(ctx context.Context, history model.TicketHistory) (*model.TicketHistory, error) {
	return u.repo.Create(ctx, history)
}

func (u *TicketHistoryUsecase) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketHistoryResponse, error) {
	log := logrus.WithField("ticket_id", ticketID)

	histories, err := u.repo.FindByTicketID(ctx, ticketID)
	if err != nil {
		log.Error("failed get histories:", err)
		return nil, err
	}

	var result []*model.TicketHistoryResponse

	for _, h := range histories {

		h.Type = mapAction(h.Action, h.FieldName)

		if h.Action == "COMMENT" || h.Action == "ONHOLD_NOTE" {
			h.Message = h.NewValue
		}

		result = append(result, h)
	}

	return result, nil
}

func mapAction(action, field string) string {
	switch action {

	case "CREATE", "CREATED":
		return "CREATED"

	case "UPDATE", "UPDATE_STATUS", "STATUS_UPDATED":
		if field == "status" {
			return "STATUS_UPDATED"
		}
		return "OTHER"

	case "COMMENT":
		return "COMMENT"

	case "ONHOLD_NOTE":
		return "ONHOLD_NOTE"
	}

	return "OTHER"
}
