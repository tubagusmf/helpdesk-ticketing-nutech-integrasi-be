package model

import (
	"context"
	"time"
)

type NotificationType string
type NotificationReferenceType string

const (
	NotificationTicketCreated  NotificationType = "TICKET_CREATED"
	NotificationTicketAssigned NotificationType = "TICKET_ASSIGNED"
	NotificationTicketUpdated  NotificationType = "TICKET_UPDATED"
	NotificationTicketComment  NotificationType = "TICKET_COMMENT"
	NotificationTicketResolved NotificationType = "TICKET_RESOLVED"
	NotificationTicketClosed   NotificationType = "TICKET_CLOSED"
)

const (
	ReferenceTicket     NotificationReferenceType = "TICKET"
	ReferenceComment    NotificationReferenceType = "COMMENT"
	ReferenceResolution NotificationReferenceType = "RESOLUTION"
)

type Notification struct {
	ID            int64                     `json:"id"`
	UserID        int64                     `json:"user_id"`
	ActorID       int64                     `json:"actor_id"`
	TicketID      int64                     `json:"ticket_id"`
	Type          NotificationType          `json:"type"`
	ReferenceType NotificationReferenceType `json:"reference_type"`
	ReferenceID   int64                     `json:"reference_id"`
	Title         string                    `json:"title"`
	Message       string                    `json:"message"`
	IsRead        bool                      `json:"is_read"`
	CreatedAt     time.Time                 `json:"created_at"`
}

type NotificationResponse struct {
	ID         int64     `json:"id"`
	TicketID   int64     `json:"ticket_id"`
	TicketCode string    `json:"ticket_code"`
	ActorName  string    `json:"actor_name"`
	Type       string    `json:"type"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateNotificationInput struct {
	UserID        int64
	ActorID       int64
	TicketID      int64
	Type          NotificationType
	ReferenceType NotificationReferenceType
	ReferenceID   int64
	Title         string
	Message       string
}

type INotificationRepository interface {
	Create(ctx context.Context, notification Notification) (*Notification, error)
	FindAllByUserID(ctx context.Context, userID int64) ([]*NotificationResponse, error)
	MarkAsRead(ctx context.Context, id int64, userID int64) error
	CountUnread(ctx context.Context, userID int64) (int64, error)
}

type INotificationUsecase interface {
	Create(ctx context.Context, in CreateNotificationInput) (*Notification, error)
	FindAllByUserID(ctx context.Context, userID int64) ([]*NotificationResponse, error)
	MarkAsRead(ctx context.Context, id int64, userID int64) error
	CountUnread(ctx context.Context, userID int64) (int64, error)
}
