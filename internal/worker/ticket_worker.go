package worker

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/config"
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
		log.Println("No staff online, retry later")

		data, _ := json.Marshal(ticket)
		config.Rdb.LPush(context.Background(), "ticket_queue", data)

		time.Sleep(10 * time.Second)
		return
	}

	type StaffLoad struct {
		User  model.User
		Count int64
		Last  *time.Time
	}

	var candidates []StaffLoad

	for _, s := range staffs {

		var count int64
		w.db.Model(&model.Ticket{}).
			Where("assigned_to_id = ? AND status IN ?", s.ID, []string{"OPEN", "IN_PROGRESS"}).
			Count(&count)

		var lastTicket model.Ticket
		err := w.db.
			Where("assigned_to_id = ?", s.ID).
			Order("created_at DESC").
			First(&lastTicket).Error

		var lastTime *time.Time
		if err == nil {
			lastTime = &lastTicket.CreatedAt
		}

		candidates = append(candidates, StaffLoad{
			User:  s,
			Count: count,
			Last:  lastTime,
		})
	}

	var filtered []StaffLoad
	now := time.Now()

	for _, c := range candidates {
		if c.Last == nil || now.Sub(*c.Last) > 2*time.Minute {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		filtered = candidates
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Count < filtered[j].Count
	})

	selected := filtered[0].User

	w.db.Model(&model.Ticket{}).
		Where("id = ?", ticket.ID).
		Update("assigned_to_id", selected.ID)

	log.Printf("Assigned ticket ID: %d to staff ID: %d (%s)", ticket.ID, selected.ID, selected.Name)
}
