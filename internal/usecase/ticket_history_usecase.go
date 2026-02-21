package usecase

import (
	"context"

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

func (u *TicketHistoryUsecase) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketHistory, error) {
	return u.repo.FindByTicketID(ctx, ticketID)
}
