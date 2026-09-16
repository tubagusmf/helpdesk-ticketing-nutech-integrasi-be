package model

import (
	"context"
	"time"
)

type TicketReassignment struct {
	ID          int64                          `json:"id"`
	TicketID    int64                          `json:"ticket_id"`
	FromUserID  *int64                         `json:"from_user_id"`
	ToUserID    int64                          `json:"to_user_id"`
	Message     string                         `json:"message"`
	CreatedAt   time.Time                      `json:"created_at"`
	Attachments []TicketReassignmentAttachment `json:"attachments"`
}

type TicketReassignmentAttachment struct {
	ID             int64     `json:"id"`
	ReassignmentID int64     `json:"reassignment_id"`
	FileName       string    `json:"file_name"`
	FileURL        string    `json:"file_url"`
	PublicID       *string   `json:"public_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type ReassignTicketInput struct {
	ToUserID int64  `json:"to_user_id" validate:"required"`
	Message  string `json:"message" validate:"required"`
}

type ITicketReassignmentRepository interface {
	Create(ctx context.Context, reassignment *TicketReassignment) error
	CreateAttachment(ctx context.Context, attachment *TicketReassignmentAttachment) error
	FindByTicketID(ctx context.Context, ticketID int64) ([]*TicketReassignment, error)
}
