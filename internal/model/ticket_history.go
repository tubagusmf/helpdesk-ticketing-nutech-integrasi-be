package model

import (
	"context"
	"time"
)

const (
	HistoryActionTicketOverdue        = "TICKET_OVERDUE"
	HistoryActionStaffResponseOverdue = "STAFF_RESPONSE_OVERDUE"
	HistoryActionStaffResponseLate    = "STAFF_RESPONSE_LATE"
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

type TicketHistoryResponse struct {
	ID          int64     `json:"id"`
	TicketID    int64     `json:"ticket_id"`
	UserID      int64     `json:"user_id"`
	UserName    string    `json:"user_name"`
	Type        string    `json:"type"`
	Action      string    `json:"action"`
	FieldName   string    `json:"field_name"`
	OldValue    *string   `json:"old_value"`
	NewValue    *string   `json:"new_value"`
	Message     *string   `json:"message"`
	OldUserName string    `json:"old_user_name"`
	NewUserName string    `json:"new_user_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type ITicketHistoryRepository interface {
	Create(ctx context.Context, history TicketHistory) (*TicketHistory, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketHistoryResponse, error)
}

type ITicketHistoryUsecase interface {
	Create(ctx context.Context, history TicketHistory) (*TicketHistory, error)
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketHistoryResponse, error)
}
