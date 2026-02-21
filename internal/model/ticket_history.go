package model

import (
	"context"
	"time"
)

type TicketHistory struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	FieldName string    `json:"field_name"`
	OldValue  *string   `json:"old_value"`
	NewValue  *string   `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

type ITicketHistoryRepository interface {
	Create(ctx context.Context, history TicketHistory) (*TicketHistory, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketHistory, error)
}

type ITicketHistoryUsecase interface {
	Create(ctx context.Context, history TicketHistory) (*TicketHistory, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketHistory, error)
}
