package usecase

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"

	ws "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"
)

type TicketCommentUsecase struct {
	repo              model.ITicketCommentRepository
	ticketHistoryRepo model.ITicketHistoryRepository
	ticketRepo        model.ITicketRepository
	wsHub             *ws.Hub
}

func commentAuthorDisplayName(user model.User) string {
	if user.RoleID == 2 && user.CodeName != nil {
		codeName := strings.TrimSpace(*user.CodeName)

		if codeName != "" {
			return codeName
		}
	}

	return user.Name
}

func NewTicketCommentUsecase(repo model.ITicketCommentRepository, ticketHistoryRepo model.ITicketHistoryRepository, ticketRepo model.ITicketRepository, wsHub *ws.Hub) model.ITicketCommentUsecase {
	return &TicketCommentUsecase{
		repo:              repo,
		ticketHistoryRepo: ticketHistoryRepo,
		ticketRepo:        ticketRepo,
		wsHub:             wsHub,
	}
}

func (u *TicketCommentUsecase) Create(ctx context.Context, comment model.TicketComment) (*model.TicketComment, error) {
	log := logrus.WithFields(logrus.Fields{
		"comment": comment,
	})

	ticket, err := u.ticketRepo.FindByID(ctx, comment.TicketID)
	if err != nil {
		log.Error("Failed find ticket:", err)
		return nil, err
	}

	if comment.UserID == ticket.ReporterID {
		comment.IsReadByUser = true
		comment.IsReadByStaff = false
		comment.IsReadByAdministrator = false
	} else {
		comment.IsReadByUser = false
		comment.IsReadByStaff = true
		comment.IsReadByAdministrator = true
	}

	result, err := u.repo.Create(ctx, comment)
	if err != nil {
		log.Error("Failed to create ticket comment: ", err)
		return nil, err
	}

	if ticket.StaffAssignedToID != nil &&
		*ticket.StaffAssignedToID == comment.UserID {

		if err := u.ticketRepo.RecordStaffFirstResponse(
			ctx,
			comment.TicketID,
			comment.UserID,
		); err != nil {
			log.Error(
				"Failed record staff first response:",
				err,
			)
		}
	}

	ws.BroadcastToRoles(
		u.wsHub,
		[]string{"ADMINISTRATOR", "STAFF", "USER", "EXECUTIVE", "ENGINEER"},
		websocket.Message{
			Type: "NEW_COMMENT",
			Data: map[string]interface{}{
				"id":         result.ID,
				"ticket_id":  result.TicketID,
				"user_id":    result.UserID,
				"user_name":  commentAuthorDisplayName(result.User),
				"message":    result.Message,
				"created_at": result.CreatedAt,
			},
		})

	message := comment.Message

	history, err := u.ticketHistoryRepo.Create(ctx, model.TicketHistory{
		TicketID: comment.TicketID,
		UserID:   comment.UserID,
		Action:   "COMMENT",
		NewValue: &message,
	})

	if err == nil {
		histories, err := u.ticketHistoryRepo.FindByTicketID(
			ctx,
			history.TicketID,
		)

		if err == nil && len(histories) > 0 {

			latest := histories[0]

			latest.Type = "COMMENT"
			latest.Message = latest.NewValue

			BroadcastTicketHistory(
				u.wsHub,
				latest,
			)
		}
	}

	messagePreview := strings.TrimSpace(comment.Message)

	runes := []rune(messagePreview)

	if len(runes) > 80 {
		messagePreview = string(runes[:80]) + "..."
	}

	publishTicketNotification(
		"ticket.comment",
		ticket,
		comment.UserID,
		model.NotificationTicketComment,
		model.ReferenceComment,
		result.ID,
		"Komentar Baru",
		"Komentar pada tiket "+
			ticket.TicketCode+
			": "+
			messagePreview,
	)

	if err != nil {
		log.Error("Failed publish notification:", err)
	}

	return result, nil
}

func (u *TicketCommentUsecase) FindByTicketID(ctx context.Context, ticketID int64) ([]*model.TicketCommentResponse, error) {
	log := logrus.WithFields(logrus.Fields{
		"ticketID": ticketID,
	})

	comment, err := u.repo.FindByTicketID(ctx, ticketID)
	if err != nil {
		log.Error("Failed to find ticket comment: ", err)
		return nil, err
	}

	return comment, nil
}

func (u *TicketCommentUsecase) MarkAsRead(ctx context.Context, ticketID int64, role string) error {
	return u.repo.MarkAsRead(ctx, ticketID, role)
}

func (u *TicketCommentUsecase) Delete(ctx context.Context, ticketID int64, commentID int64, userID int64) error {
	log := logrus.WithFields(logrus.Fields{
		"ticket_id":  ticketID,
		"comment_id": commentID,
		"user_id":    userID,
	})

	err := u.repo.DeleteByID(
		ctx,
		ticketID,
		commentID,
		userID,
	)
	if err != nil {
		log.Error("Failed to delete ticket comment: ", err)
		return err
	}

	ws.BroadcastToRoles(
		u.wsHub,
		[]string{
			"ADMINISTRATOR",
			"STAFF",
			"USER",
			"EXECUTIVE",
			"ENGINEER",
		},
		websocket.Message{
			Type: websocket.EventCommentDeleted,
			Data: map[string]interface{}{
				"id":        commentID,
				"ticket_id": ticketID,
				"user_id":   userID,
			},
		},
	)

	return nil
}
