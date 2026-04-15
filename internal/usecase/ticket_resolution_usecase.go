package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketResolutionUsecase struct {
	db             *gorm.DB
	resolutionRepo model.ITicketResolutionRepository
	historyRepo    model.ITicketHistoryRepository
	ticketRepo     model.ITicketRepository
}

func NewTicketResolutionUsecase(
	db *gorm.DB,
	resolutionRepo model.ITicketResolutionRepository,
	historyRepo model.ITicketHistoryRepository,
	ticketRepo model.ITicketRepository,
) model.ITicketResolutionUsecase {
	return &TicketResolutionUsecase{
		db:             db,
		resolutionRepo: resolutionRepo,
		historyRepo:    historyRepo,
		ticketRepo:     ticketRepo,
	}
}

func (u *TicketResolutionUsecase) Create(ctx context.Context, userID int64, in model.CreateTicketResolutionInput) (*model.TicketResolution, error) {
	if in.Status != model.StatusResolved {
		return nil, errors.New("resolution only allowed for RESOLVED status")
	}

	var ticket model.Ticket
	if err := u.db.Where("id = ?", in.TicketID).First(&ticket).Error; err != nil {
		return nil, err
	}

	oldStatus := ticket.Status

	if ticket.PausedAt != nil {
		duration := time.Since(*ticket.PausedAt).Seconds()
		ticket.TotalPaused += int64(duration)
		ticket.PausedAt = nil
	}

	completionTime := in.CompletionTime
	if completionTime.IsZero() {
		completionTime = time.Now()
	}

	resolution := model.TicketResolution{
		TicketID:        in.TicketID,
		CauseID:         in.CauseID,
		SolutionID:      in.SolutionID,
		ResolutionNotes: in.ResolutionNotes,
		CompletionTime:  completionTime,
		AttachmentURL:   in.AttachmentURL,
	}

	tx := u.db.Begin()

	createdResolution, err := u.resolutionRepo.Create(ctx, tx, resolution)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Model(&model.Ticket{}).
		Where("id = ?", ticket.ID).
		Updates(map[string]interface{}{
			"status":       model.StatusResolved,
			"paused_at":    nil,
			"total_paused": ticket.TotalPaused,
		}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	oldStatusStr := string(oldStatus)
	newStatusStr := string(model.StatusResolved)

	history := model.TicketHistory{
		TicketID:  ticket.ID,
		UserID:    userID,
		Action:    "UPDATE_STATUS",
		FieldName: "status",
		OldValue:  &oldStatusStr,
		NewValue:  &newStatusStr,
	}

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return createdResolution, nil
}

func (u *TicketResolutionUsecase) FindByTicketID(ctx context.Context, ticketID int64) (*model.TicketResolution, error) {
	return u.resolutionRepo.FindByTicketID(ctx, ticketID)
}

func (u *TicketResolutionUsecase) UpdateStatus(ctx context.Context, ticketID int64, userID int64, in model.UpdateTicketStatusInput) error {
	log := logrus.WithFields(logrus.Fields{
		"ticket_id": ticketID,
		"user_id":   userID,
		"status":    in.Status,
	})

	var ticket model.Ticket
	if err := u.db.WithContext(ctx).
		Where("id = ?", ticketID).
		First(&ticket).Error; err != nil {
		return err
	}

	tx := u.db.Begin()

	oldStatus := ticket.Status

	if err := tx.Model(&model.Ticket{}).
		Where("id = ?", ticket.ID).
		Updates(map[string]interface{}{
			"status": in.Status,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	oldStatusStr := string(oldStatus)
	newStatusStr := string(in.Status)

	history := model.TicketHistory{
		TicketID:  ticket.ID,
		UserID:    userID,
		Action:    "UPDATE_STATUS",
		FieldName: "status",
		OldValue:  &oldStatusStr,
		NewValue:  &newStatusStr,
	}

	log.Infof("INSERT HISTORY: %+v", history)

	if err := tx.Create(&history).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}
