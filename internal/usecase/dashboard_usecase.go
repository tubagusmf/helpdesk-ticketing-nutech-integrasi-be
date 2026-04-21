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

func (u *DashboardUsecase) GetSummary(ctx context.Context, filter map[string]interface{}) (*model.DashboardSummary, error) {
	return u.repo.GetSummary(ctx, filter)
}

func (u *DashboardUsecase) GetStatus(ctx context.Context, filter map[string]interface{}) (*model.StatusDistribution, error) {
	return u.repo.GetStatusDistribution(ctx, filter)
}

func (u *DashboardUsecase) GetPriority(ctx context.Context, filter map[string]interface{}) ([]model.PriorityDistribution, error) {
	return u.repo.GetPriorityDistribution(ctx, filter)
}

func (u *DashboardUsecase) GetVolume(ctx context.Context, filter map[string]interface{}) ([]model.VolumeProject, error) {
	return u.repo.GetVolumeProject(ctx, filter)
}
