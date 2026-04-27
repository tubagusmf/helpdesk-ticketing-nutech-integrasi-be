package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketUsecase struct {
	ticketRepo        model.ITicketRepository
	ticketHistoryRepo model.ITicketHistoryRepository
	db                *gorm.DB
}

func NewTicketUsecase(
	db *gorm.DB,
	ticketRepo model.ITicketRepository,
	historyRepo model.ITicketHistoryRepository,
) model.ITicketUsecase {
	return &TicketUsecase{
		db:                db,
		ticketRepo:        ticketRepo,
		ticketHistoryRepo: historyRepo,
	}
}

func (u *TicketUsecase) FindAll(ctx context.Context, filter model.Ticket, search string, startDate string, endDate string, page int, limit int) ([]*model.TicketResponse, int64, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"search": search,
	})

	tickets, total, err := u.ticketRepo.FindAll(ctx, filter, search, startDate, endDate, page, limit)
	if err != nil {
		log.Error("Failed to fetch tickets: ", err)
		return nil, 0, err
	}

	return tickets, total, nil
}

func (u *TicketUsecase) FindByID(ctx context.Context, id int64) (*model.Ticket, error) {
	ticket, err := u.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (u *TicketUsecase) Create(ctx context.Context, reporterID int64, in model.CreateTicketInput, attachmentPath *string) (*model.Ticket, bool, error) {
	if err := validate.Struct(in); err != nil {
		return nil, false, err
	}

	var count int64
	u.db.Model(&model.User{}).
		Where("role_id = 2 AND is_online = true").
		Count(&count)

	isAssigned := count > 0

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	var dueAt time.Time
	switch in.Priority {
	case model.PriorityUrgent:
		dueAt = now.Add(15 * time.Minute)
	case model.PriorityHigh:
		dueAt = now.Add(1 * time.Hour)
	case model.PriorityMedium:
		dueAt = now.Add(2 * time.Hour)
	case model.PriorityLow:
		dueAt = now.Add(4 * time.Hour)
	}

	ticket := model.Ticket{
		TicketCode:   generateTicketCode(),
		ProjectID:    in.ProjectID,
		LocationID:   in.LocationID,
		PartID:       in.PartID,
		AssetID:      in.AssetID,
		ReporterID:   reporterID,
		Priority:     in.Priority,
		Description:  in.Description,
		DueAt:        dueAt,
		Status:       model.StatusOpen,
		Attachment:   attachmentPath,
		AssignedToID: nil,
	}

	if err := u.db.WithContext(ctx).Create(&ticket).Error; err != nil {
		return nil, false, err
	}

	_, err := u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
		TicketID: ticket.ID,
		UserID:   reporterID,
		Action:   "CREATED",
	})

	if err != nil {
		logrus.Error("failed insert CREATED history:", err)
	}

	data, _ := json.Marshal(ticket)
	config.Rdb.LPush(config.Ctx(), "ticket_queue", data)

	return &ticket, isAssigned, nil
}

func (u *TicketUsecase) UpdateStatus(ctx context.Context, id int64, userID int64, in model.UpdateTicketStatusInput) error {
	log := logrus.WithFields(logrus.Fields{
		"ticket_id": id,
		"user_id":   userID,
		"status":    in.Status,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("validation error:", err)
		return err
	}

	ticket, err := u.ticketRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("ticket not found:", err)
		return err
	}

	oldStatus := ticket.Status

	now := time.Now()

	if in.Status == model.StatusOnHold && ticket.PausedAt == nil {
		ticket.PausedAt = &now
	}

	if in.Status == model.StatusOnHold {
		ticket.OnholdNotes = &in.OnholdNotes
	} else {
		ticket.OnholdNotes = nil
	}

	if ticket.PausedAt != nil &&
		(in.Status == model.StatusOpen ||
			in.Status == model.StatusInProgress ||
			in.Status == model.StatusResolved) {

		duration := time.Since(*ticket.PausedAt).Seconds()
		ticket.TotalPaused += int64(duration)
		ticket.PausedAt = nil
	}

	ticket.Status = in.Status

	if err := u.ticketRepo.Update(ctx, *ticket); err != nil {
		log.Error("failed update ticket:", err)
		return err
	}

	oldStatusStr := string(oldStatus)
	newStatusStr := string(in.Status)

	_, err = u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
		TicketID:  id,
		UserID:    userID,
		Action:    "STATUS_UPDATED",
		FieldName: "status",
		OldValue:  &oldStatusStr,
		NewValue:  &newStatusStr,
	})

	if err != nil {
		log.Error("failed insert STATUS history:", err)
		return err
	}

	if in.Status == model.StatusOnHold && in.OnholdNotes != "" {
		notes := in.OnholdNotes

		_, err := u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
			TicketID: id,
			UserID:   userID,
			Action:   "ONHOLD_NOTE",
			NewValue: &notes,
		})

		if err != nil {
			log.Error("failed insert ONHOLD_NOTE:", err)
		} else {
			log.Info("success insert ONHOLD_NOTE")
		}
	}

	return nil
}

func (u *TicketUsecase) Delete(ctx context.Context, id int64) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	if err := u.ticketRepo.Delete(ctx, id); err != nil {
		log.Error("Failed to delete ticket: ", err)
		return err
	}

	return nil
}

func isValidStatus(status model.TicketStatus) bool {
	switch status {
	case model.StatusOpen,
		model.StatusInProgress,
		model.StatusResolved,
		model.StatusClosed,
		model.StatusOnHold:
		return true
	default:
		return false
	}
}

func generateTicketCode() string {
	now := time.Now()
	return fmt.Sprintf("TCK-%d%02d%02d-%d",
		now.Year(),
		now.Month(),
		now.Day(),
		now.Unix()%10000,
	)
}

func (u *TicketUsecase) validateStaff(userID int64) error {
	var user model.User

	if err := u.db.First(&user, userID).Error; err != nil {
		return err
	}

	if user.RoleID != 2 {
		return errors.New("assigned user must be STAFF")
	}

	return nil
}
