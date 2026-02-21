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

func (u *TicketResolutionUsecase) Create(
	ctx context.Context,
	userID int64,
	in model.CreateTicketResolutionInput,
) (*model.TicketResolution, error) {

	log := logrus.WithFields(logrus.Fields{"in": in})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	var ticket model.Ticket
	if err := u.db.WithContext(ctx).
		Where("id = ?", in.TicketID).
		First(&ticket).Error; err != nil {
		return nil, err
	}

	if ticket.Status == "RESOLVED" {
		return nil, errors.New("ticket already resolved")
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

	oldStatus := ticket.Status
	newStatus := "RESOLVED"

	if err := tx.Model(&model.Ticket{}).
		Where("id = ?", in.TicketID).
		Update("status", newStatus).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	oldStatusStr := string(oldStatus)
	newStatusStr := string(newStatus)

	history := model.TicketHistory{
		TicketID:  in.TicketID,
		UserID:    userID,
		Action:    "UPDATE_STATUS",
		FieldName: "status",
		OldValue:  &oldStatusStr,
		NewValue:  &newStatusStr,
		CreatedAt: time.Now(),
	}

	if err := tx.WithContext(ctx).Create(&history).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return createdResolution, nil
}

func (u *TicketResolutionUsecase) FindByTicketID(ctx context.Context, ticketID int64) (*model.TicketResolution, error) {
	return u.resolutionRepo.FindByTicketID(ctx, ticketID)
}
