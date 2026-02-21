package model

import (
	"context"
	"time"
)

type TicketResolution struct {
	ID              int64     `json:"id"`
	TicketID        int64     `json:"ticket_id"`
	CauseID         int64     `json:"cause_id"`
	SolutionID      int64     `json:"solution_id"`
	ResolutionNotes string    `json:"resolution_notes"`
	CompletionTime  time.Time `json:"completion_time"`
	AttachmentURL   string    `json:"attachment_url"`
	CreatedAt       time.Time `json:"created_at"`

	Cause    *Cause    `json:"cause,omitempty"`
	Solution *Solution `json:"solution,omitempty"`
}

type CreateTicketResolutionInput struct {
	TicketID        int64     `json:"ticket_id" validate:"required"`
	CauseID         int64     `json:"cause_id" validate:"required"`
	SolutionID      int64     `json:"solution_id" validate:"required"`
	ResolutionNotes string    `json:"resolution_notes"`
	CompletionTime  time.Time `json:"completion_time"`
	AttachmentURL   string    `json:"attachment_url"`
}

type ITicketResolutionRepository interface {
	Create(ctx context.Context, tx interface{}, resolution TicketResolution) (*TicketResolution, error)
	FindByTicketID(ctx context.Context, ticketID int64) (*TicketResolution, error)
}

type ITicketResolutionUsecase interface {
	Create(ctx context.Context, userID int64, in CreateTicketResolutionInput) (*TicketResolution, error)
	FindByTicketID(ctx context.Context, ticketID int64) (*TicketResolution, error)
}
