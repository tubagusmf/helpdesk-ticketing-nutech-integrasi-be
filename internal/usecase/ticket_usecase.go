package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

type TicketUsecase struct {
	ticketRepo model.ITicketRepository
}

func NewTicketUsecase(ticketRepo model.ITicketRepository) model.ITicketUsecase {
	return &TicketUsecase{
		ticketRepo: ticketRepo,
	}
}

func (u *TicketUsecase) FindAll(ctx context.Context, filter model.Ticket) ([]*model.Ticket, error) {
	log := logrus.WithFields(logrus.Fields{
		"filter": filter,
	})

	tickets, err := u.ticketRepo.FindAll(ctx, filter)
	if err != nil {
		log.Error("Failed to fetch tickets: ", err)
		return nil, err
	}

	return tickets, nil
}

func (u *TicketUsecase) FindByID(ctx context.Context, id int64) (*model.Ticket, error) {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	ticket, err := u.ticketRepo.FindByID(ctx, id)
	if err != nil {
		log.Error("Failed to find ticket: ", err)
		return nil, err
	}

	return ticket, nil
}

func (u *TicketUsecase) Create(ctx context.Context, reporterID int64, in model.CreateTicketInput) (*model.Ticket, error) {
	log := logrus.WithFields(logrus.Fields{
		"in": in,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return nil, err
	}

	if in.DueAt.Before(time.Now()) {
		return nil, errors.New("due date cannot be in the past")
	}

	ticket := model.Ticket{
		TicketCode:   in.TicketCode,
		ProjectID:    in.ProjectID,
		LocationID:   in.LocationID,
		PartID:       in.PartID,
		AssetID:      in.AssetID,
		ReporterID:   reporterID,
		AssignedToID: in.AssignedToID,
		Priority:     in.Priority,
		Description:  in.Description,
		DueAt:        in.DueAt,
	}

	created, err := u.ticketRepo.Create(ctx, ticket)
	if err != nil {
		log.Error("Failed to create ticket: ", err)
		return nil, err
	}

	return created, nil
}

func (u *TicketUsecase) UpdateStatus(ctx context.Context, id int64, in model.UpdateTicketStatusInput) error {
	log := logrus.WithFields(logrus.Fields{
		"id": id,
	})

	if err := validate.Struct(in); err != nil {
		log.Error("Validation error: ", err)
		return err
	}

	if !isValidStatus(in.Status) {
		return errors.New("invalid ticket status")
	}

	ticket, err := u.ticketRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	ticket.Status = in.Status

	if in.Status == model.StatusResolved {
		now := time.Now()
		ticket.ResolvedAt = &now
	}

	if err := u.ticketRepo.Update(ctx, *ticket); err != nil {
		log.Error("Failed to update ticket: ", err)
		return err
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
