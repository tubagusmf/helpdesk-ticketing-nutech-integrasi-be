package model

import "context"

type DashboardSummary struct {
	TotalTicket       int64   `json:"total_ticket"`
	SLABreach         int64   `json:"sla_breach"`
	TicketSelesai     int64   `json:"ticket_selesai"`
	TicketOnHold      int64   `json:"ticket_onhold"`
	AvgResolutionTime float64 `json:"avg_resolution_time"`
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

type IDashboardRepository interface {
	GetSummary(ctx context.Context, filter map[string]interface{}) (*DashboardSummary, error)
	GetStatusDistribution(ctx context.Context, filter map[string]interface{}) (*StatusDistribution, error)
	GetPriorityDistribution(ctx context.Context, filter map[string]interface{}) ([]PriorityDistribution, error)
	GetVolumeProject(ctx context.Context, filter map[string]interface{}) ([]VolumeProject, error)
}
