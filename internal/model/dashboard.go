package model

import (
	"context"
	"time"
)

type DashboardSummary struct {
	TotalTicket       int64   `json:"total_ticket"`
	TicketOpen        int64   `json:"ticket_open"`
	TicketOnHold      int64   `json:"ticket_onhold"`
	TicketResolved    int64   `json:"ticket_resolved"`
	TicketClosed      int64   `json:"ticket_closed"`
	SLABreach         int64   `json:"sla_breach"`
	AvgResolutionTime float64 `json:"avg_resolution_time"`
}

type DashboardFilter struct {
	ProjectID int64  `json:"project_id"`
	PartID    int64  `json:"part_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	UserID    int64  `json:"user_id"`
	Role      string `json:"role"`
}

type DashboardProject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type StatusDistribution struct {
	Open       int64 `json:"open"`
	InProgress int64 `json:"in_progress"`
	Resolved   int64 `json:"resolved"`
	Closed     int64 `json:"closed"`
	OnHold     int64 `json:"onhold"`
}

type PriorityDistribution struct {
	Priority string `json:"priority"`
	Total    int64  `json:"total"`
}

type VolumeProject struct {
	Project string `json:"project"`
	Total   int64  `json:"total"`
}

type DashboardProjectVolume struct {
	ProjectID int64  `json:"project_id"`
	Project   string `json:"project"`
	Total     int64  `json:"total"`
}

type DashboardLocationVolume struct {
	LocationID int64  `json:"location_id"`
	Location   string `json:"location"`
	Total      int64  `json:"total"`
}

type DashboardTrend struct {
	Date  string `json:"date"`
	Total int64  `json:"total"`
}

type DashboardTicket struct {
	ID          int64     `json:"id"`
	TicketCode  string    `json:"ticket_code"`
	Description string    `json:"description"`
	Project     string    `json:"project"`
	Location    string    `json:"location"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type DashboardProjectSummary struct {
	ProjectID  int64  `json:"project_id"`
	Project    string `json:"project"`
	Total      int64  `json:"total"`
	Open       int64  `json:"open"`
	InProgress int64  `json:"in_progress"`
	OnHold     int64  `json:"onhold"`
	Resolved   int64  `json:"resolved"`
	Closed     int64  `json:"closed"`
}

type DashboardStaffSummary struct {
	StaffID    int64  `json:"staff_id"`
	StaffName  string `json:"staff_name"`
	Total      int64  `json:"total"`
	Open       int64  `json:"open"`
	InProgress int64  `json:"in_progress"`
	OnHold     int64  `json:"onhold"`
	Resolved   int64  `json:"resolved"`
	Closed     int64  `json:"closed"`
}

type IDashboardRepository interface {
	GetSummary(ctx context.Context, filter DashboardFilter) (*DashboardSummary, error)
	GetStatusDistribution(ctx context.Context, filter DashboardFilter) (*StatusDistribution, error)
	GetPriorityDistribution(ctx context.Context, filter DashboardFilter) ([]PriorityDistribution, error)
	GetVolumeProject(ctx context.Context, filter DashboardFilter) ([]VolumeProject, error)
	GetProjects(ctx context.Context, filter DashboardFilter) ([]DashboardProject, error)
	GetTopProjects(ctx context.Context, filter DashboardFilter) ([]DashboardProjectVolume, error)
	GetTopLocations(ctx context.Context, filter DashboardFilter) ([]DashboardLocationVolume, error)
	GetIncidentTrend(ctx context.Context, filter DashboardFilter) ([]DashboardTrend, error)
	GetOpenTickets(ctx context.Context, filter DashboardFilter) ([]DashboardTicket, error)
	GetOnHoldTickets(ctx context.Context, filter DashboardFilter) ([]DashboardTicket, error)
	GetProjectSummary(ctx context.Context, filter DashboardFilter) ([]DashboardProjectSummary, error)
	GetStaffSummary(ctx context.Context, filter DashboardFilter) ([]DashboardStaffSummary, error)
}
