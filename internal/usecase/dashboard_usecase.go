package usecase

import (
	"context"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type DashboardUsecase struct {
	repo model.IDashboardRepository
}

func NewDashboardUsecase(repo model.IDashboardRepository) *DashboardUsecase {
	return &DashboardUsecase{repo: repo}
}

func (u *DashboardUsecase) GetSummary(ctx context.Context, filter model.DashboardFilter) (*model.DashboardSummary, error) {
	return u.repo.GetSummary(ctx, filter)
}

func (u *DashboardUsecase) GetStatus(ctx context.Context, filter model.DashboardFilter) (*model.StatusDistribution, error) {
	return u.repo.GetStatusDistribution(ctx, filter)
}

func (u *DashboardUsecase) GetPriority(ctx context.Context, filter model.DashboardFilter) ([]model.PriorityDistribution, error) {
	return u.repo.GetPriorityDistribution(ctx, filter)
}

func (u *DashboardUsecase) GetVolume(ctx context.Context, filter model.DashboardFilter) ([]model.VolumeProject, error) {
	return u.repo.GetVolumeProject(ctx, filter)
}

func (u *DashboardUsecase) GetProjects(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardProject, error) {
	return u.repo.GetProjects(ctx, filter)
}

func (u *DashboardUsecase) GetTopProjects(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardProjectVolume, error) {
	return u.repo.GetTopProjects(ctx, filter)
}

func (u *DashboardUsecase) GetTopLocations(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardLocationVolume, error) {
	return u.repo.GetTopLocations(ctx, filter)
}

func (u *DashboardUsecase) GetIncidentTrend(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardTrend, error) {
	return u.repo.GetIncidentTrend(ctx, filter)
}

func (u *DashboardUsecase) GetOpenTickets(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardTicket, error) {
	return u.repo.GetOpenTickets(ctx, filter)
}

func (u *DashboardUsecase) GetOnHoldTickets(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardTicket, error) {
	return u.repo.GetOnHoldTickets(ctx, filter)
}

func (u *DashboardUsecase) GetProjectSummary(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardProjectSummary, error) {
	return u.repo.GetProjectSummary(ctx, filter)
}

func (u *DashboardUsecase) GetStaffSummary(ctx context.Context, filter model.DashboardFilter) ([]model.DashboardStaffSummary, error) {
	return u.repo.GetStaffSummary(ctx, filter)
}
