package model

import (
	"context"
	"time"
)

type TicketEngineerResolution struct {
	ID          int64                                `json:"id"`
	TicketID    int64                                `json:"ticket_id"`
	EngineerID  int64                                `json:"engineer_id"`
	Solution    string                               `json:"solution"`
	CreatedAt   time.Time                            `json:"created_at"`
	UpdatedAt   time.Time                            `json:"updated_at"`
	Attachments []TicketEngineerResolutionAttachment `json:"attachments" gorm:"foreignKey:ResolutionID;references:ID"`
}

type TicketEngineerResolutionAttachment struct {
	ID           int64     `json:"id"`
	ResolutionID int64     `json:"resolution_id"`
	FileName     string    `json:"file_name"`
	FileURL      string    `json:"file_url"`
	PublicID     *string   `json:"public_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateEngineerResolutionInput struct {
	Solution string `json:"solution" validate:"required"`
}

type ITicketEngineerResolutionRepository interface {
	Create(ctx context.Context, resolution *TicketEngineerResolution) error
	CreateAttachment(ctx context.Context, attachment *TicketEngineerResolutionAttachment) error
	FindByTicketID(ctx context.Context, ticketID int64) (*TicketEngineerResolution, error)
}
