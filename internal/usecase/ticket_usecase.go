package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	ws "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TicketUsecase struct {
	ticketRepo         model.ITicketRepository
	ticketHistoryRepo  model.ITicketHistoryRepository
	ticketReassignRepo model.ITicketReassignmentRepository
	projectRepo        model.IProjectRepository
	ticketCommentRepo  model.ITicketCommentRepository
	db                 *gorm.DB
	hub                *ws.Hub
}

func NewTicketUsecase(
	db *gorm.DB,
	ticketRepo model.ITicketRepository,
	historyRepo model.ITicketHistoryRepository,
	reassignRepo model.ITicketReassignmentRepository,
	projectRepo model.IProjectRepository,
	ticketCommentRepo model.ITicketCommentRepository,
	hub *ws.Hub,
) model.ITicketUsecase {
	return &TicketUsecase{
		db:                 db,
		ticketRepo:         ticketRepo,
		ticketHistoryRepo:  historyRepo,
		ticketReassignRepo: reassignRepo,
		projectRepo:        projectRepo,
		ticketCommentRepo:  ticketCommentRepo,
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
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	ticket, err := u.ticketRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to fetch ticket: ", err)
		return nil, err
	}

	return ticket, nil
}

func (u *TicketUsecase) FindResponseByID(ctx context.Context, id int64) (*model.TicketResponse, error) {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	ticket, err := u.ticketRepo.FindResponseByID(ctx, id)
	if err != nil {
		log.Error("Failed to fetch ticket response: ", err)
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

	if in.Status == model.StatusCancel {
		return errors.New(
			"gunakan endpoint pembatalan tiket untuk mengubah status menjadi CANCEL",
		)
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

	if ticket.Status == model.StatusCancel {
		return errors.New(
			"tiket yang sudah dibatalkan tidak dapat diubah statusnya",
		)
	}

	if err := u.ticketRepo.RecordStaffFirstResponse(
		ctx,
		id,
		userID,
	); err != nil {
		log.Error("failed record staff first response:", err)
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
		model.StatusResolved,
		model.StatusClosed,
		model.StatusOnHold,
		model.StatusCancel:
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
		Status:     model.TicketReassignmentPending,
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

	err = helper.PublishNotificationEvent(
		"ticket.reassigned",
		model.NotificationEvent{
			EventType:     "TICKET_REASSIGNED",
			UserID:        in.ToUserID,
			ActorID:       userID,
			TicketID:      ticketID,
			TicketCode:    ticket.TicketCode,
			ReferenceType: "TICKET",
			ReferenceID:   ticketID,
			Title:         "Tiket Di-reassign",
			Message:       "No Tiket: " + ticket.TicketCode + " | Tiket telah di-reassign kepada Anda.",
		},
	)

	if err != nil {
		log.Error("failed publish reassignment notification:", err)
	}

	return nil
}

func (u *TicketUsecase) GetReassignmentByTicketID(ctx context.Context, ticketID int64) (*model.TicketReassignment, error) {
	reassignments, err := u.ticketReassignRepo.FindByTicketID(
		ctx,
		ticketID,
	)

	if err != nil {
		return nil, err
	}

	if len(reassignments) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	data := reassignments[0]

	for _, reassignment := range reassignments {
		if reassignment == nil {
			continue
		}

		if data == nil || reassignment.CreatedAt.After(data.CreatedAt) {
			data = reassignment
		}
	}

	if data == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return data, nil
}

func (u *TicketUsecase) ResponseTicket(ctx context.Context, ticketID int64, userID int64) error {
	ticket, err := u.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return err
	}

	if ticket.StaffAssignedToID == nil {
		return errors.New("ticket belum memiliki staff")
	}

	if *ticket.StaffAssignedToID != userID {
		return errors.New("ticket bukan tanggung jawab staff ini")
	}

	if ticket.StaffFirstResponseAt != nil {
		return errors.New("ticket sudah diresponse")
	}

	if err := u.ticketRepo.RecordStaffFirstResponse(
		ctx,
		ticketID,
		userID,
	); err != nil {
		return err
	}

	const responseMessage = "Baik akan ditindak lanjut"

	comment := model.TicketComment{
		TicketID:              ticketID,
		UserID:                userID,
		Message:               responseMessage,
		IsReadByUser:          false,
		IsReadByStaff:         true,
		IsReadByAdministrator: false,
	}

	result, err := u.ticketCommentRepo.Create(ctx, comment)
	if err != nil {
		return err
	}

	ws.BroadcastToRoles(
		u.hub,
		[]string{
			"ADMINISTRATOR",
			"STAFF",
			"USER",
		},
		ws.Message{
			Type: "NEW_COMMENT",
			Data: map[string]interface{}{
				"id":         result.ID,
				"ticket_id":  result.TicketID,
				"user_id":    result.UserID,
				"user_name":  result.User.Name,
				"message":    result.Message,
				"created_at": result.CreatedAt,
			},
		},
	)

	err = helper.PublishNotificationEvent(
		"ticket.comment",
		model.NotificationEvent{
			EventType:     string(model.NotificationTicketComment),
			UserID:        ticket.ReporterID,
			ActorID:       userID,
			TicketID:      ticketID,
			TicketCode:    ticket.TicketCode,
			ReferenceType: string(model.ReferenceComment),
			ReferenceID:   result.ID,
			Title:         "Tiket Direspons",
			Message: "No Tiket: " + ticket.TicketCode +
				" | Staff CCIT telah merespons tiket Anda: \"" +
				result.Message + "\"",
		},
	)

	if err != nil {
		logrus.Error(
			"failed publish response notification:",
			err,
		)
	}

	message := responseMessage

	history, err := u.ticketHistoryRepo.Create(
		ctx,
		model.TicketHistory{
			TicketID: ticketID,
			UserID:   userID,
			Action:   "COMMENT",
			NewValue: &message,
		},
	)

	if err != nil {
		return err
	}

	histories, err := u.ticketHistoryRepo.FindByTicketID(
		ctx,
		history.TicketID,
	)

	if err == nil && len(histories) > 0 {
		latest := histories[0]

		latest.Type = "COMMENT"
		latest.Message = latest.NewValue

		BroadcastTicketHistory(
			u.hub,
			latest,
		)
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
			"STAFF",
			"ADMINISTRATOR",
			"USER",
		},
		ws.Message{
			Type: ws.EventTicketUpdated,
			Data: ticketResp,
		},
	)

	return nil
}

func (u *TicketUsecase) ExportCustom(ctx context.Context, req model.CustomTicketExportRequest, role string, userID int64) ([]*model.TicketExportRow, error) {
	switch role {
	case "STAFF", "ADMINISTRATOR", "EXECUTIVE":
	default:
		return nil, errors.New(
			"only STAFF, ADMINISTRATOR, and EXECUTIVE can export custom ticket",
		)
	}

	if len(req.Columns) == 0 {
		return nil, errors.New(
			"at least one export column is required",
		)
	}

	var assignedToID *int64

	if req.AssignedToID != 0 {
		assignedToID = &req.AssignedToID
	}

	filter := model.Ticket{
		TicketCode:   req.TicketCode,
		ProjectID:    req.ProjectID,
		AssignedToID: assignedToID,
		ReporterID:   req.ReporterID,
		Priority:     req.Priority,
		Status:       req.Status,
	}

	return u.ticketRepo.FindForCustomExport(
		ctx,
		filter,
		req.Search,
		req.StartDate,
		req.EndDate,
		role,
		userID,
	)
}

func (u *TicketUsecase) ReassignToOnlineStaff(ctx context.Context, ticketID int64, userID int64, justification string) (*model.TicketResponse, error) {
	justification = strings.TrimSpace(justification)

	if justification == "" {
		return nil, errors.New("justifikasi reassign wajib diisi")
	}

	if len([]rune(justification)) > 1000 {
		return nil, errors.New("justifikasi maksimal 1000 karakter")
	}

	now := time.Now()

	var ticketCode string
	var selectedStaff model.User
	var assignmentHistory model.TicketHistory
	var reasonHistory model.TicketHistory

	err := u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var actor model.User

		if err := tx.
			Select("id", "role_id").
			First(&actor, userID).Error; err != nil {
			return err
		}

		if actor.RoleID != 2 {
			return errors.New("fitur ini hanya untuk STAFF")
		}

		var ticket model.Ticket

		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND deleted_at IS NULL", ticketID).
			First(&ticket).Error; err != nil {
			return err
		}

		if ticket.Status == model.StatusResolved ||
			ticket.Status == model.StatusClosed {
			return errors.New(
				"ticket yang sudah selesai tidak dapat direassign",
			)
		}

		var currentStaffID int64

		if ticket.StaffAssignedToID != nil {
			currentStaffID = *ticket.StaffAssignedToID
		} else if ticket.AssignedToID != nil {
			currentStaffID = *ticket.AssignedToID
		}

		if currentStaffID != userID {
			return errors.New(
				"ticket bukan tanggung jawab Staff ini",
			)
		}

		var candidates []model.User

		if err := tx.
			Where(
				"role_id = ? AND is_online = ? AND is_active = ?",
				2,
				true,
				true,
			).
			Where("id <> ?", userID).
			Order("last_ticket_assigned_at ASC NULLS FIRST").
			Order("id ASC").
			Find(&candidates).Error; err != nil {
			return err
		}

		if len(candidates) == 0 {
			return errors.New(
				"tidak ada Staff lain yang sedang online",
			)
		}

		cutoff := now.Add(-time.Minute)
		found := false

		for _, candidate := range candidates {
			if candidate.LastTicketAssignedAt == nil ||
				!candidate.LastTicketAssignedAt.After(cutoff) {
				selectedStaff = candidate
				found = true
				break
			}
		}

		if !found {
			return errors.New(
				"semua Staff online lain masih dalam cooldown " +
					"assignment 1 menit. Silakan coba kembali",
			)
		}

		ticketCode = ticket.TicketCode

		updates := map[string]interface{}{
			"staff_assigned_to_id":        selectedStaff.ID,
			"staff_assigned_at":           now,
			"staff_first_response_at":     nil,
			"staff_response_time_seconds": nil,
			"updated_at":                  now,
		}

		if ticket.AssignedToID == nil ||
			*ticket.AssignedToID == userID {
			updates["assigned_to_id"] = selectedStaff.ID
		}

		result := tx.
			Model(&model.Ticket{}).
			Where("id = ?", ticketID).
			Updates(updates)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("gagal memperbarui assignment ticket")
		}

		if err := tx.
			Model(&model.User{}).
			Where("id = ?", selectedStaff.ID).
			Update("last_ticket_assigned_at", now).Error; err != nil {
			return err
		}

		oldValue := fmt.Sprintf("%d", currentStaffID)
		newValue := fmt.Sprintf("%d", selectedStaff.ID)

		assignmentHistory = model.TicketHistory{
			TicketID:  ticketID,
			UserID:    userID,
			Action:    "STAFF_REASSIGNED",
			FieldName: "staff_assigned_to_id",
			OldValue:  &oldValue,
			NewValue:  &newValue,
		}

		if err := tx.Create(&assignmentHistory).Error; err != nil {
			return err
		}

		reasonHistory = model.TicketHistory{
			TicketID:  ticketID,
			UserID:    userID,
			Action:    "STAFF_REASSIGNMENT_REASON",
			FieldName: "justification",
			NewValue:  &justification,
		}

		if err := tx.Create(&reasonHistory).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	histories, historyErr := u.ticketHistoryRepo.FindByTicketID(
		ctx,
		ticketID,
	)

	if historyErr != nil {
		logrus.WithError(historyErr).Error(
			"failed to fetch staff reassignment history",
		)
	} else {
		for _, history := range histories {
			if history == nil {
				continue
			}

			switch history.ID {
			case assignmentHistory.ID:
				history.Type = "STAFF_REASSIGNED"

				BroadcastTicketHistory(
					u.hub,
					history,
				)

			case reasonHistory.ID:
				history.Type = "STAFF_REASSIGNMENT_REASON"
				history.Message = history.NewValue

				BroadcastTicketHistory(
					u.hub,
					history,
				)
			}
		}
	}

	ticketResp, err := u.ticketRepo.FindResponseByID(ctx, ticketID)
	if err != nil {
		return nil, err
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

	err = helper.PublishNotificationEvent(
		"ticket.reassigned",
		model.NotificationEvent{
			EventType:     "TICKET_REASSIGNED",
			UserID:        selectedStaff.ID,
			ActorID:       userID,
			TicketID:      ticketID,
			TicketCode:    ticketCode,
			ReferenceType: "TICKET",
			ReferenceID:   ticketID,
			Title:         "Tiket Dialihkan ke Anda",
			Message: "No Tiket: " + ticketCode +
				" | Justifikasi: " + justification,
		},
	)

	if err != nil {
		logrus.Error(
			"failed publish staff reassignment notification:",
			err,
		)
	}

	return ticketResp, nil
}

func (u *TicketUsecase) CancelTicket(ctx context.Context, id int64, userID int64, role string, in model.CancelTicketInput) error {
	role = strings.ToUpper(strings.TrimSpace(role))

	if role != "STAFF" && role != "ADMINISTRATOR" {
		return errors.New(
			"hanya Staff CCIT dan Administrator yang dapat membatalkan tiket",
		)
	}

	reason := strings.TrimSpace(in.Reason)

	if reason == "" {
		return errors.New("alasan pembatalan tiket wajib diisi")
	}

	if len(reason) > 1000 {
		return errors.New("alasan pembatalan maksimal 1000 karakter")
	}

	ticket, err := u.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if ticket.Status == model.StatusClosed {
		return errors.New("tiket yang sudah CLOSED tidak dapat dibatalkan")
	}

	if ticket.Status == model.StatusCancel {
		return errors.New("tiket sudah dibatalkan")
	}

	totalPaused := ticket.TotalPaused

	if ticket.PausedAt != nil {
		duration := time.Since(*ticket.PausedAt).Seconds()

		if duration > 0 {
			totalPaused += int64(duration)
		}
	}

	oldStatus := string(ticket.Status)
	newStatus := string(model.StatusCancel)

	if err := u.ticketRepo.CancelTicket(ctx, id, userID, reason, totalPaused); err != nil {
		return err
	}

	statusHistory, historyErr := u.ticketHistoryRepo.Create(
		ctx,
		model.TicketHistory{
			TicketID:  id,
			UserID:    userID,
			Action:    "STATUS_UPDATED",
			FieldName: "status",
			OldValue:  &oldStatus,
			NewValue:  &newStatus,
		},
	)

	if historyErr != nil {
		logrus.WithError(historyErr).Error(
			"gagal menyimpan history status pembatalan tiket",
		)
	} else if statusHistory != nil {
		histories, findErr :=
			u.ticketHistoryRepo.FindByTicketID(ctx, id)

		if findErr == nil && len(histories) > 0 {
			latest := histories[0]
			latest.Type = "STATUS_UPDATED"

			BroadcastTicketHistory(u.hub, latest)
		}
	}

	reasonHistory, reasonHistoryErr := u.ticketHistoryRepo.Create(
		ctx,
		model.TicketHistory{
			TicketID:  id,
			UserID:    userID,
			Action:    "CANCELLATION_REASON",
			FieldName: "cancellation_reason",
			NewValue:  &reason,
		},
	)

	if reasonHistoryErr != nil {
		logrus.WithError(reasonHistoryErr).Error(
			"gagal menyimpan history alasan pembatalan tiket",
		)
	} else if reasonHistory != nil {
		histories, findErr :=
			u.ticketHistoryRepo.FindByTicketID(ctx, id)

		if findErr == nil && len(histories) > 0 {
			latest := histories[0]
			latest.Type = "CANCELLATION_REASON"
			latest.Message = latest.NewValue

			BroadcastTicketHistory(u.hub, latest)
		}
	}

	ticketResp, err := u.ticketRepo.FindResponseByID(ctx, id)
	if err != nil {
		return err
	}

	ws.BroadcastToRoles(
		u.hub,
		[]string{
			"STAFF",
			"ADMINISTRATOR",
			"USER",
			"ENGINEER",
			"EXECUTIVE",
		},
		ws.Message{
			Type: ws.EventTicketStatusUpdate,
			Data: ticketResp,
		},
	)

	recipients := []int64{
		ticket.ReporterID,
	}

	if ticket.StaffAssignedToID != nil {
		recipients = append(
			recipients,
			*ticket.StaffAssignedToID,
		)
	}

	if ticket.AssignedToID != nil {
		recipients = append(
			recipients,
			*ticket.AssignedToID,
		)
	}

	seenRecipients := make(map[int64]struct{})

	for _, recipientID := range recipients {
		if recipientID <= 0 || recipientID == userID {
			continue
		}

		if _, exists := seenRecipients[recipientID]; exists {
			continue
		}

		seenRecipients[recipientID] = struct{}{}

		notificationErr := helper.PublishNotificationEvent(
			"ticket.cancelled",
			model.NotificationEvent{
				EventType:     string(model.NotificationTicketCancel),
				UserID:        recipientID,
				ActorID:       userID,
				TicketID:      id,
				TicketCode:    ticket.TicketCode,
				ReferenceType: string(model.ReferenceTicket),
				ReferenceID:   id,
				Title:         "Tiket Dibatalkan",
				Message: "No Tiket: " + ticket.TicketCode +
					" | Tiket dibatalkan. Alasan: " + reason,
			},
		)

		if notificationErr != nil {
			logrus.WithFields(logrus.Fields{
				"ticket_id":    id,
				"recipient_id": recipientID,
			}).WithError(notificationErr).Error(
				"gagal mengirim notifikasi pembatalan tiket",
			)
		}
	}

	return nil
}
