package worker

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketWorker struct {
	db *gorm.DB
}

func NewTicketWorker(db *gorm.DB) *TicketWorker {
	return &TicketWorker{db: db}
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

		if now.Sub(*c.Last) >= 2*time.Minute {
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
			Message:       "Kamu Menerima Tiket " + ticket.TicketCode,
		},
	)

	if err != nil {
		log.Println("failed publish notification:", err)
	}

}
