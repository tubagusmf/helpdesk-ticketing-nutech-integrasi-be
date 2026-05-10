package model

import (
	"context"
	"time"
)

type TicketComment struct {
	ID                    int64     `json:"id"`
	TicketID              int64     `json:"ticket_id"`
	UserID                int64     `json:"user_id"`
	Message               string    `json:"message"`
	IsReadByUser          bool      `json:"is_read_by_user"`
	IsReadByStaff         bool      `json:"is_read_by_staff"`
	IsReadByAdministrator bool      `json:"is_read_by_administrator"`
	CreatedAt             time.Time `json:"created_at"`
}

type TicketCommentResponse struct {
	ID        int64     `json:"id"`
	TicketID  int64     `json:"ticket_id"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type ITicketCommentRepository interface {
	Create(ctx context.Context, comment TicketComment) (*TicketComment, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketCommentResponse, error)
	CountUnreadByTicket(ctx context.Context, ticketID int64, role string, userID int64) (int64, error)
	MarkAsRead(ctx context.Context, ticketID int64, role string) error
}

type ITicketCommentUsecase interface {
	Create(ctx context.Context, comment TicketComment) (*TicketComment, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketCommentResponse, error)
	MarkAsRead(ctx context.Context, ticketID int64, role string) error
}
