package worker

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	ws "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"
	"gorm.io/gorm"
)

type TicketWorker struct {
	db  *gorm.DB
	hub *ws.Hub
}

func NewTicketWorker(db *gorm.DB, hub *ws.Hub) *TicketWorker {
	return &TicketWorker{
		db:  db,
		hub: hub,
	}
}

func (w *TicketWorker) Start() {
	go w.monitorSLAHistory()
	go w.monitorStaleUsersOffline()

	for {
		result, err := config.Rdb.BRPop(
			context.Background(),
			0,
			"ticket_queue",
		).Result()

		if err != nil {
			log.Println("Redis error:", err)
			continue
		}

		var ticket model.Ticket

		if err := json.Unmarshal(
			[]byte(result[1]),
			&ticket,
		); err != nil {
			log.Println("Invalid ticket queue payload:", err)
			continue
		}

		w.process(ticket)
	}
}

func (w *TicketWorker) monitorSLAHistory() {
	w.checkSLAHistory()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.checkSLAHistory()
	}
}

func (w *TicketWorker) checkSLAHistory() {
	now := time.Now()

	var tickets []model.Ticket

	err := w.db.
		Where("deleted_at IS NULL").
		Where("status = ?", model.StatusOpen).
		Where(`
            due_at <= ?
            OR (
                staff_assigned_at IS NOT NULL
                AND staff_first_response_at IS NULL
                AND staff_assigned_at <= ?
            )
        `,
			now,
			now.Add(-3*time.Minute),
		).
		Find(&tickets).Error

	if err != nil {
		log.Println("[SLA HISTORY] failed load tickets:", err)
		return
	}

	for _, ticket := range tickets {
		actorID := ticket.ReporterID

		if ticket.StaffAssignedToID != nil {
			actorID = *ticket.StaffAssignedToID
		} else if ticket.AssignedToID != nil {
			actorID = *ticket.AssignedToID
		}

		if !ticket.DueAt.After(now) {
			message := "Tiket melewati batas penyelesaian SLA."

			if err := w.recordSLAHistory(
				ticket,
				actorID,
				model.HistoryActionTicketOverdue,
				nil,
				ticket.DueAt,
				message,
			); err != nil {
				log.Println(
					"[SLA HISTORY] overdue:",
					ticket.ID,
					err,
				)
			}
		}

		if ticket.StaffAssignedToID == nil ||
			ticket.StaffAssignedAt == nil ||
			ticket.StaffFirstResponseAt != nil {
			continue
		}

		responseDeadline := ticket.StaffAssignedAt.Add(
			3 * time.Minute,
		)

		if now.Before(responseDeadline) {
			continue
		}

		assignmentKey := ticket.StaffAssignedAt.
			UTC().
			Format(time.RFC3339Nano)

		message := "Tiket belum direspons setelah melewati batas SLA respons 3 menit."

		if err := w.recordSLAHistory(
			ticket,
			actorID,
			model.HistoryActionStaffResponseOverdue,
			&assignmentKey,
			responseDeadline,
			message,
		); err != nil {
			log.Println(
				"[SLA HISTORY] response overdue:",
				ticket.ID,
				err,
			)
		}
	}
}

func (w *TicketWorker) monitorStaleUsersOffline() {
	w.markStaleUsersOffline()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		w.markStaleUsersOffline()
	}
}

func (w *TicketWorker) markStaleUsersOffline() {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	cutoff := time.Now().Add(-30 * time.Second)

	result := w.db.WithContext(ctx).
		Model(&model.User{}).
		Where(`
        is_online = ?
        AND deleted_at IS NULL
        AND (
            last_heartbeat_at IS NULL
            OR last_heartbeat_at < ?
        )
    `, true, cutoff).
		Updates(map[string]interface{}{
			"is_online":  false,
			"last_seen":  time.Now(),
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		log.Printf(
			"[ONLINE STATUS] gagal menandai user Offline: %v",
			result.Error,
		)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf(
			"[ONLINE STATUS] %d user ditandai Offline karena heartbeat kedaluwarsa",
			result.RowsAffected,
		)
	}
}

func (w *TicketWorker) recordSLAHistory(ticket model.Ticket, actorID int64, action string, oldValue *string, occurredAt time.Time, message string) error {
	ctx := context.Background()

	query := w.db.WithContext(ctx).
		Model(&model.TicketHistory{}).
		Where("ticket_id = ?", ticket.ID).
		Where("action = ?", action).
		Where("field_name = ?", "sla")

	if oldValue == nil {
		query = query.Where("old_value IS NULL")
	} else {
		query = query.Where("old_value = ?", *oldValue)
	}

	var count int64

	if err := query.Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	history := model.TicketHistory{
		TicketID:  ticket.ID,
		UserID:    actorID,
		Action:    action,
		FieldName: "sla",
		OldValue:  oldValue,
		NewValue:  &message,
		CreatedAt: occurredAt,
	}

	if err := w.db.WithContext(ctx).
		Create(&history).Error; err != nil {
		return err
	}

	response := model.TicketHistoryResponse{
		ID:        history.ID,
		TicketID:  history.TicketID,
		UserID:    history.UserID,
		UserName:  "System",
		Type:      action,
		Action:    action,
		FieldName: history.FieldName,
		OldValue:  history.OldValue,
		NewValue:  history.NewValue,
		Message:   history.NewValue,
		CreatedAt: history.CreatedAt,
	}

	ws.BroadcastToRoles(
		w.hub,
		[]string{
			"ADMINISTRATOR",
			"STAFF",
			"USER",
			"EXECUTIVE",
			"ENGINEER",
		},
		ws.Message{
			Type: "TICKET_HISTORY",
			Data: response,
		},
	)

	log.Printf(
		"[SLA HISTORY] ticket=%d action=%s",
		ticket.ID,
		action,
	)

	return nil
}

func (w *TicketWorker) process(ticket model.Ticket) {
	var currentTicket model.Ticket

	err := w.db.
		Select("id, assigned_to_id").
		First(&currentTicket, ticket.ID).Error

	if err != nil {
		log.Printf(
			"[WORKER] failed to load ticket %d: %v",
			ticket.ID,
			err,
		)
		return
	}

	if currentTicket.AssignedToID != nil {
		log.Printf(
			"[WORKER] ticket %d already assigned to %d, skip",
			ticket.ID,
			*currentTicket.AssignedToID,
		)
		return
	}

	var staffs []model.User

	cutoff := time.Now().Add(-30 * time.Second)

	err = w.db.
		Where(`
        role_id = ?
        AND is_active = ?
        AND is_online = ?
        AND deleted_at IS NULL
        AND last_heartbeat_at IS NOT NULL
        AND last_heartbeat_at >= ?
    `,
			2,
			true,
			true,
			cutoff,
		).
		Find(&staffs).Error

	if err != nil {
		log.Printf(
			"[WORKER] gagal mengambil kandidat Staff CCIT: %v",
			err,
		)

		data, marshalErr := json.Marshal(ticket)
		if marshalErr != nil {
			log.Printf(
				"[WORKER] gagal encode tiket %d: %v",
				ticket.ID,
				marshalErr,
			)
			return
		}

		if queueErr := config.Rdb.LPush(
			context.Background(),
			"ticket_queue",
			data,
		).Err(); queueErr != nil {
			log.Printf(
				"[WORKER] gagal memasukkan kembali tiket %d: %v",
				ticket.ID,
				queueErr,
			)
		}

		return
	}

	if len(staffs) == 0 {

		log.Println("[WORKER] no staff online, retry later")

		data, _ := json.Marshal(ticket)

		time.Sleep(10 * time.Second)

		config.Rdb.LPush(
			context.Background(),
			"ticket_queue",
			data,
		)

		return
	}

	type StaffLoad struct {
		User model.User
		Last *time.Time
	}

	var candidates []StaffLoad

	for _, s := range staffs {
		lastTime := s.LastTicketAssignedAt
		candidates = append(candidates, StaffLoad{
			User: s,
			Last: lastTime,
		})
	}

	var available []StaffLoad
	now := time.Now()

	for _, c := range candidates {

		if c.Last != nil {
			log.Printf(
				"[WORKER] staff=%s last_assign=%v diff=%v",
				c.User.Name,
				*c.Last,
				now.Sub(*c.Last),
			)
		}

		if c.Last == nil {
			available = append(available, c)
			continue
		}

		if now.Sub(*c.Last) >= 1*time.Minute {
			available = append(available, c)
		}
	}

	if len(available) == 0 {

		log.Println("[WORKER] all staff still in cooldown, retry later")

		data, _ := json.Marshal(ticket)

		time.Sleep(10 * time.Second)

		config.Rdb.LPush(
			context.Background(),
			"ticket_queue",
			data,
		)

		return
	}

	sort.Slice(available, func(i, j int) bool {

		if available[i].Last == nil {
			return true
		}

		if available[j].Last == nil {
			return false
		}

		if available[i].Last.Equal(*available[j].Last) {
			return available[i].User.ID < available[j].User.ID
		}

		return available[i].Last.Before(*available[j].Last)
	})

	selected := available[0].User

	nowAssign := time.Now()

	result := w.db.Model(&model.Ticket{}).
		Where("id = ? AND assigned_to_id IS NULL", ticket.ID).
		Updates(map[string]interface{}{
			"assigned_to_id":              selected.ID,
			"staff_assigned_to_id":        selected.ID,
			"staff_assigned_at":           nowAssign,
			"staff_first_response_at":     nil,
			"staff_response_time_seconds": nil,
			"updated_at":                  nowAssign,
		})

	if result.Error != nil {
		log.Println("failed assign ticket:", result.Error)
		return
	}

	if result.RowsAffected == 0 {
		log.Printf(
			"[WORKER] ticket %d already assigned by another process",
			ticket.ID,
		)
		return
	}

	err = w.db.Model(&model.User{}).
		Where("id = ?", selected.ID).
		Update("last_ticket_assigned_at", nowAssign).Error

	if err != nil {
		log.Println("failed update last_ticket_assigned_at:", err)
	}

	w.db.Preload("Reporter").First(&ticket, ticket.ID)

	var updatedTicket model.TicketResponse

	err = w.db.
		Table("tickets").
		Select(`
		tickets.id,
		tickets.ticket_code,
		tickets.priority,
		tickets.status,
		tickets.description,
		tickets.created_at,
		tickets.due_at,
		tickets.reporter_id,
		tickets.part_id,
		tickets.asset_id,
		tickets.attachment,
		tickets.assigned_to_id,

		tickets.staff_assigned_to_id,
        tickets.staff_assigned_at,
        tickets.staff_first_response_at,
        tickets.staff_response_time_seconds,
		reporter.name AS reporter_name,

		CASE
			WHEN assigned.role_id = 2
			AND NULLIF(BTRIM(assigned.code_name), '') IS NOT NULL
			THEN assigned.code_name
			ELSE assigned.name
		END AS assigned_to_name,

		CASE
			WHEN staff_assigned.role_id = 2
			AND NULLIF(BTRIM(staff_assigned.code_name), '') IS NOT NULL
			THEN staff_assigned.code_name
			ELSE staff_assigned.name
		END AS staff_assigned_to_name,

		projects.name AS project_name,
		locations.name AS location_name,
		parts.name AS part_name,
		asset_ids.name AS asset_code
	`).
		Joins(`
		LEFT JOIN users reporter
		ON reporter.id = tickets.reporter_id
	`).
		Joins(`
		LEFT JOIN users assigned
		ON assigned.id = tickets.assigned_to_id
	`).
		Joins(`
		LEFT JOIN users staff_assigned
		ON staff_assigned.id = tickets.staff_assigned_to_id
	`).
		Joins(`
		LEFT JOIN projects
		ON projects.id = tickets.project_id
	`).
		Joins(`
		LEFT JOIN locations
		ON locations.id = tickets.location_id
	`).
		Joins(`
		LEFT JOIN parts
		ON parts.id = tickets.part_id
	`).
		Joins(`
		LEFT JOIN asset_ids
		ON asset_ids.id = tickets.asset_id
	`).
		Where("tickets.id = ?", ticket.ID).
		Scan(&updatedTicket).Error

	if err != nil {
		logrus.Error(
			"failed load updated ticket:",
			err,
		)
		return
	}

	log.Printf(
		"[WORKER] ticket assigned ticket_id=%d assigned_to=%d",
		ticket.ID,
		selected.ID,
	)

	message := ws.Message{
		Type: ws.EventNewTicket,
		Data: updatedTicket,
	}

	payload, _ := json.Marshal(message)

	go func() {
		w.hub.BroadcastToRole <- ws.RoleMessage{
			Role:    "STAFF",
			Message: payload,
		}
	}()

	go func() {
		w.hub.BroadcastToRole <- ws.RoleMessage{
			Role:    "ADMINISTRATOR",
			Message: payload,
		}
	}()

	go func() {
		w.hub.BroadcastToRole <- ws.RoleMessage{
			Role:    "USER",
			Message: payload,
		}
	}()

	log.Printf(
		"[WORKER] websocket broadcast sent ticket_id=%d",
		ticket.ID,
	)

	err = helper.PublishNotificationEvent(
		"ticket.created",
		model.NotificationEvent{
			EventType:     "TICKET_ASSIGNED",
			UserID:        selected.ID,
			ActorID:       ticket.ReporterID,
			TicketID:      ticket.ID,
			TicketCode:    ticket.TicketCode,
			ReferenceType: "TICKET",
			ReferenceID:   ticket.ID,
			Title:         "Tiket Masuk",
			Message:       "No Tiket: " + ticket.TicketCode + " | Pelapor: " + ticket.Reporter.Name,
		},
	)

	if err != nil {
		log.Println("failed publish notification:", err)
	}

}
