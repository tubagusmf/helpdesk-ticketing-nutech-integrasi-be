package model

import (
	"context"
	"time"
)

type TicketStatus string
type TicketPriority string

const (
	StatusOpen       TicketStatus = "OPEN"
	StatusInProgress TicketStatus = "IN_PROGRESS"
	StatusResolved   TicketStatus = "RESOLVED"
	StatusClosed     TicketStatus = "CLOSED"
	StatusOnHold     TicketStatus = "ONHOLD"

	PriorityLow    TicketPriority = "LOW"
	PriorityMedium TicketPriority = "MEDIUM"
	PriorityHigh   TicketPriority = "HIGH"
	PriorityUrgent TicketPriority = "URGENT"
)

type Ticket struct {
	ID           int64          `json:"id"`
	TicketCode   string         `json:"ticket_code"`
	ProjectID    int64          `json:"project_id"`
	LocationID   int64          `json:"location_id"`
	PartID       int64          `json:"part_id"`
	AssetID      int64          `json:"asset_id"`
	ReporterID   int64          `json:"reporter_id"`
	AssignedToID int64          `json:"assigned_to_id"`
	Status       TicketStatus   `json:"status"`
	Priority     TicketPriority `json:"priority"`
	Description  string         `json:"description"`
	Attachment   *string        `json:"attachment"`
	DueAt        time.Time      `json:"due_at"`
	ResolvedAt   *time.Time     `json:"resolved_at"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}

type TicketResponse struct {
	ID          int64  `json:"id"`
	TicketCode  string `json:"ticket_code"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	Description string `json:"description"`

	ProjectName  string `json:"project_name"`
	LocationName string `json:"location_name"`
	AssetCode    string `json:"asset_code"`

	ReporterName   string `json:"reporter_name"`
	AssignedToName string `json:"assigned_to_name"`

	CreatedAt time.Time `json:"created_at"`
	DueAt     time.Time `json:"due_at"`
}

type CreateTicketInput struct {
	ProjectID    int64          `json:"project_id" validate:"required"`
	LocationID   int64          `json:"location_id" validate:"required"`
	PartID       int64          `json:"part_id" validate:"required"`
	AssetID      int64          `json:"asset_id" validate:"required"`
	AssignedToID int64          `json:"assigned_to_id" validate:"required"`
	Priority     TicketPriority `json:"priority" validate:"required"`
	Description  string         `json:"description" validate:"required"`
}

type UpdateTicketStatusInput struct {
	Status TicketStatus `json:"status" validate:"required"`
}

type ITicketRepository interface {
	FindAll(ctx context.Context, filter Ticket) ([]*TicketResponse, error)
	FindByID(ctx context.Context, id int64) (*Ticket, error)
	Create(ctx context.Context, ticket Ticket) (*Ticket, error)
	Update(ctx context.Context, ticket Ticket) error
	Delete(ctx context.Context, id int64) error
}

type ITicketUsecase interface {
	FindAll(ctx context.Context, filter Ticket) ([]*TicketResponse, error)
	FindByID(ctx context.Context, id int64) (*Ticket, error)
	Create(ctx context.Context, reporterID int64, in CreateTicketInput, attachmentPath *string) (*Ticket, error)
	UpdateStatus(ctx context.Context, id int64, userID int64, in UpdateTicketStatusInput) error
	Delete(ctx context.Context, id int64) error
}
