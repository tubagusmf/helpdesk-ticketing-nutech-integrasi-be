package model

type NotificationEvent struct {
	EventType string `json:"event_type"`

	UserID  int64 `json:"user_id"`
	ActorID int64 `json:"actor_id"`

	TicketID   int64  `json:"ticket_id"`
	TicketCode string `json:"ticket_code"`

	ReferenceType string `json:"reference_type"`
	ReferenceID   int64  `json:"reference_id"`

	Title   string `json:"title"`
	Message string `json:"message"`
}
