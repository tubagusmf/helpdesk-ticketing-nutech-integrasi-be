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
	ws "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"
	"gorm.io/gorm"
)

type TicketUsecase struct {
	ticketRepo         model.ITicketRepository
	ticketHistoryRepo  model.ITicketHistoryRepository
	ticketReassignRepo model.ITicketReassignmentRepository
	projectRepo        model.IProjectRepository
	db                 *gorm.DB
	hub                *ws.Hub
}

func NewTicketUsecase(
	db *gorm.DB,
	ticketRepo model.ITicketRepository,
	historyRepo model.ITicketHistoryRepository,
	reassignRepo model.ITicketReassignmentRepository,
	projectRepo model.IProjectRepository,
	hub *ws.Hub,
) model.ITicketUsecase {
	return &TicketUsecase{
		db:                 db,
		ticketRepo:         ticketRepo,
		ticketHistoryRepo:  historyRepo,
		ticketReassignRepo: reassignRepo,
		projectRepo:        projectRepo,
		hub:                hub,
	}
}

func (u *TicketUsecase) FindAll(ctx context.Context, filter model.Ticket, search string, startDate string, endDate string, page int, limit int, role string, userID int64) ([]*model.TicketResponse, int64, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
		"search": search,
	})

	tickets, total, err := u.ticketRepo.FindAll(ctx, filter, search, startDate, endDate, page, limit, role, userID)
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

	isAssigned := false

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

	project, err := u.projectRepo.FindByID(ctx, in.ProjectID)
	if err != nil {
		return nil, false, err
	}

	seq, err := u.getNextTicketSequence(ctx, project.CodePrefix, in.ProjectID)
	if err != nil {
		return nil, false, err
	}

	ticketCode := generateTicketCode(project.CodePrefix, int(seq))

	ticket := model.Ticket{
		TicketCode:   ticketCode,
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

	if err := u.db.WithContext(ctx).
		First(&ticket, ticket.ID).Error; err != nil {
		return nil, false, err
	}

	history, err := u.ticketHistoryRepo.Create(
		ctx,
		model.TicketHistory{
			TicketID: ticket.ID,
			UserID:   reporterID,
			Action:   "CREATED",
		},
	)

	if err == nil {
		histories, err := u.ticketHistoryRepo.FindByTicketID(
			ctx,
			history.TicketID,
		)

		if err == nil && len(histories) > 0 {

			latest := histories[0]

			latest.Type = "CREATED"

			BroadcastTicketHistory(
				u.hub,
				latest,
			)
		}
	}

	if ticket.AssignedToID != nil {
		logrus.Infof(
			"ticket saved assigned_to_id=%d",
			*ticket.AssignedToID,
		)
	} else {
		logrus.Infof(
			"ticket saved with no assigned staff yet",
		)
	}

	if err != nil {
		logrus.Error("failed insert CREATED history:", err)
	}

	ticketResp, err := u.ticketRepo.FindResponseByID(ctx, ticket.ID)
	if err != nil {
		return nil, false, err
	}

	data, _ := json.Marshal(ticketResp)
	config.Rdb.LPush(config.Ctx(), "ticket_queue", data)

	ws.BroadcastToRoles(
		u.hub,
		[]string{
			"STAFF",
			"ADMINISTRATOR",
		},
		ws.Message{
			Type: ws.EventTicketUpdated,
			Data: ticketResp,
		},
	)

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

	history, err := u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
		TicketID:  id,
		UserID:    userID,
		Action:    "STATUS_UPDATED",
		FieldName: "status",
		OldValue:  &oldStatusStr,
		NewValue:  &newStatusStr,
	})

	if err == nil {
		histories, err := u.ticketHistoryRepo.FindByTicketID(
			ctx,
			history.TicketID,
		)

		if err == nil && len(histories) > 0 {

			for _, h := range histories {
				logrus.Infof(
					"[WS HISTORY] status history id=%d old=%v new=%v",
					h.ID,
					h.OldValue,
					h.NewValue,
				)
			}

			latest := histories[0]

			latest.Type = "STATUS_UPDATED"

			payload, _ := json.Marshal(latest)

			logrus.Infof(
				"[STATUS HISTORY PAYLOAD] %s",
				string(payload),
			)

			BroadcastTicketHistory(
				u.hub,
				latest,
			)
		}
	}

	if err != nil {
		log.Error("failed insert STATUS history:", err)
		return err
	}

	if in.Status == model.StatusOnHold && in.OnholdNotes != "" {
		notes := in.OnholdNotes

		history, err := u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
			TicketID: id,
			UserID:   userID,
			Action:   "ONHOLD_NOTE",
			NewValue: &notes,
		})

		if err == nil {

			histories, err := u.ticketHistoryRepo.FindByTicketID(
				ctx,
				history.TicketID,
			)

			if err == nil && len(histories) > 0 {

				latest := histories[0]

				latest.Type = "ONHOLD_NOTE"
				latest.Message = latest.NewValue

				BroadcastTicketHistory(
					u.hub,
					latest,
				)
			}
		}

		if err != nil {
			log.Error("failed insert ONHOLD_NOTE:", err)
		} else {
			log.Info("success insert ONHOLD_NOTE")
		}
	}

	ticketResp, err := u.ticketRepo.FindResponseByID(ctx, id)
	if err != nil {
		log.Error("failed fetch updated ticket response:", err)
		return err
	}

	ws.BroadcastToRoles(
		u.hub,
		[]string{
			"STAFF",
			"ADMINISTRATOR",
			"USER",
		},
		ws.Message{
			Type: ws.EventTicketStatusUpdate,
			Data: ticketResp,
		},
	)

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

func generateTicketCode(prefix string, seq int) string {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	return fmt.Sprintf("%s-%04d%02d%02d-%04d",
		prefix,
		now.Year(),
		now.Month(),
		now.Day(),
		seq,
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

func (u *TicketUsecase) getNextTicketSequence(ctx context.Context, projectCode string, projectID int64) (int64, error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	date := now.Format("20060102")
	key := fmt.Sprintf("ticket:%s:%s", projectCode, date)

	seq, err := config.Rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	tomorrow := now.Add(24 * time.Hour).Truncate(24 * time.Hour)
	ttl := time.Until(tomorrow)
	config.Rdb.Expire(ctx, key, ttl)

	if seq == 1 {
		count, err := u.ticketRepo.CountByProjectToday(ctx, projectID)
		if err == nil && count > 0 {
			seq = count + 1
			config.Rdb.Set(ctx, key, seq, ttl)
		}
	}

	return seq, nil
}

func (u *TicketUsecase) Reassign(ctx context.Context, ticketID int64, userID int64, in model.ReassignTicketInput, attachments []model.TicketReassignmentAttachment) error {
	log := logrus.WithFields(logrus.Fields{
		"ticket_id":  ticketID,
		"user_id":    userID,
		"to_user_id": in.ToUserID,
		"message":    in.Message,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("validation error:", err)
		return err
	}

	ticket, err := u.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		log.Error("Failed to find ticket:", err)
		return err
	}

	var targetUser model.User

	if err := u.db.WithContext(ctx).
		Where("id = ?", in.ToUserID).
		First(&targetUser).Error; err != nil {
		return err
	}

	if targetUser.RoleID != 5 {
		return errors.New("ticket can only be reassigned to ENGINEER")
	}

	var count int64

	err = u.db.WithContext(ctx).
		Table("user_projects").
		Where("user_id = ?", in.ToUserID).
		Where("project_id = ?", ticket.ProjectID).
		Count(&count).
		Error

	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New(
			"engineer is not assigned to this project",
		)
	}

	oldAssignedID := ticket.AssignedToID
	newAssignedID := in.ToUserID

	ticket.AssignedToID = &newAssignedID

	if err := u.ticketRepo.Update(ctx, *ticket); err != nil {
		return err
	}

	reassignment := &model.TicketReassignment{
		TicketID:   ticketID,
		FromUserID: oldAssignedID,
		ToUserID:   in.ToUserID,
		Message:    in.Message,
	}

	if err := u.ticketReassignRepo.Create(
		ctx,
		reassignment,
	); err != nil {
		return err
	}

	for _, attachment := range attachments {

		attachment.ReassignmentID = reassignment.ID

		if err := u.ticketReassignRepo.CreateAttachment(
			ctx,
			&attachment,
		); err != nil {
			return err
		}
	}

	oldValue := ""
	if oldAssignedID != nil {
		oldValue = fmt.Sprintf("%d", *oldAssignedID)
	}

	newValue := fmt.Sprintf("%d", newAssignedID)

	history, err := u.ticketHistoryRepo.Create(
		ctx,
		model.TicketHistory{
			TicketID:  ticketID,
			UserID:    userID,
			Action:    "REASSIGNED",
			FieldName: "assigned_to_id",
			OldValue:  &oldValue,
			NewValue:  &newValue,
		},
	)

	if err == nil {

		histories, err := u.ticketHistoryRepo.FindByTicketID(
			ctx,
			history.TicketID,
		)

		if err == nil && len(histories) > 0 {

			latest := histories[0]
			latest.Type = "REASSIGNED"

			BroadcastTicketHistory(
				u.hub,
				latest,
			)
		}
	}

	ticketResp, err := u.ticketRepo.FindResponseByID(
		ctx,
		ticketID,
	)

	if err != nil {
		return err
	}

	ws.BroadcastToRoles(
		u.hub,
		[]string{
			"ADMINISTRATOR",
			"EXECUTIVE",
			"STAFF",
			"USER",
			"ENGINEER",
		},
		ws.Message{
			Type: ws.EventTicketUpdated,
			Data: ticketResp,
		},
	)

	return nil
}
