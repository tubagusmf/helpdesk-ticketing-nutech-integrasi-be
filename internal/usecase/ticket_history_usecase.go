package usecase

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	ws "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"
)

type TicketHistoryUsecase struct {
	repo model.ITicketHistoryRepository
	hub  *ws.Hub
}

func NewTicketHistoryUsecase(repo model.ITicketHistoryRepository, hub *ws.Hub) model.ITicketHistoryUsecase {
	return &TicketHistoryUsecase{
		repo: repo,
		hub:  hub,
	}
}

func (u *TicketHistoryUsecase) Create(ctx context.Context, history model.TicketHistory) (*model.TicketHistory, error) {
	return u.repo.Create(ctx, history)
}

func (u *TicketHistoryUsecase) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketHistoryResponse, error) {
	log := logrus.WithField("ticket_id", ticketID)

	histories, err := u.repo.FindByTicketID(ctx, ticketID)
	if err != nil {
		log.Error("failed get histories:", err)
		return nil, err
	}

	var result []*model.TicketHistoryResponse

	for _, h := range histories {

		h.Type = mapAction(h.Action, h.FieldName)

		if h.Action == "COMMENT" || h.Action == "ONHOLD_NOTE" {
			h.Message = h.NewValue
		}

		result = append(result, h)
	}

	return result, nil
}

func (u *TicketHistoryUsecase) BroadcastLatestHistory(ctx context.Context, ticketID int64) {
	histories, err := u.FindByTicketID(ctx, ticketID)

	if err != nil {
		logrus.Error(err)
		return
	}

	if len(histories) == 0 {
		return
	}

	BroadcastTicketHistory(
		u.hub,
		histories[0],
	)

	logrus.Infof(
		"[BROADCAST HISTORY] ticket=%d history=%d",
		ticketID,
		histories[0].ID,
	)
}

func mapAction(action, field string) string {
	switch action {

	case "CREATE", "CREATED":
		return "CREATED"

	case "UPDATE", "UPDATE_STATUS", "STATUS_UPDATED":
		if field == "status" {
			return "STATUS_UPDATED"
		}
		return "OTHER"

	case "COMMENT":
		return "COMMENT"

	case "ONHOLD_NOTE":
		return "ONHOLD_NOTE"
	}

	return "OTHER"
}

func BroadcastTicketHistory(hub *ws.Hub, history interface{}) {
	logrus.Infof(
		"[WS SEND HISTORY] %+v",
		history,
	)

	msg := ws.Message{
		Type: "TICKET_HISTORY",
		Data: history,
	}

	payload, _ := json.Marshal(msg)

	roles := []string{
		"ADMINISTRATOR",
		"STAFF",
		"USER",
	}

	for _, role := range roles {

		hub.BroadcastToRole <- ws.RoleMessage{
			Role:    role,
			Message: payload,
		}
	}
}
