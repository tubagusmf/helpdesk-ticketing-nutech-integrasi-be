package model

import (
	"context"
	"time"
)

type TicketComment struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	UserID    int64     `json:"user_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type ITicketCommentRepository interface {
	Create(ctx context.Context, comment TicketComment) (*TicketComment, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketComment, error)
}

type ITicketCommentUsecase interface {
	Create(ctx context.Context, comment TicketComment) (*TicketComment, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketComment, error)
}
