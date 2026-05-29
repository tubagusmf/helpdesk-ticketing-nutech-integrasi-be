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

func NewTicketWorker(
	db *gorm.DB,
	hub *ws.Hub,
) *TicketWorker {
	return &TicketWorker{
		db:  db,
		hub: hub,
	}
}

func (w *TicketWorker) Start() {
	for {
		result, err := config.Rdb.BRPop(context.Background(), 0, "ticket_queue").Result()
		if err != nil {
			log.Println("Redis error:", err)
			continue
		}

		var ticket model.Ticket
		json.Unmarshal([]byte(result[1]), &ticket)

		w.process(ticket)
	}
}

func (w *TicketWorker) process(ticket model.Ticket) {
	var staffs []model.User
	w.db.Where("role_id = 2 AND is_online = true").Find(&staffs)

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

	result := w.db.Model(&model.Ticket{}).
		Where("id = ? AND assigned_to_id IS NULL", ticket.ID).
		Update("assigned_to_id", selected.ID)

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

	nowAssign := time.Now()

	err := w.db.Model(&model.User{}).
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

		reporter.name as reporter_name,
		assigned.name as assigned_to_name,
		projects.name as project_name,
		locations.name as location_name,
		parts.name as part_name,
		asset_ids.name as asset_code
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
