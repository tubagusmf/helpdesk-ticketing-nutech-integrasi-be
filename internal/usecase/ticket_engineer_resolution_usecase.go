package usecase

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/helper"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	ws "github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/websocket"
	"gorm.io/gorm"
)

const (
	MaxEngineerResolutionFileSize = 10 * 1024 * 1024
)

type TicketEngineerResolutionUsecase struct {
	db                *gorm.DB
	resolutionRepo    model.ITicketEngineerResolutionRepository
	ticketRepo        model.ITicketRepository
	userRepo          model.IUserRepository
	ticketHistoryRepo model.ITicketHistoryRepository
	hub               *ws.Hub
}

func NewTicketEngineerResolutionUsecase(
	db *gorm.DB,
	resolutionRepo model.ITicketEngineerResolutionRepository,
	ticketRepo model.ITicketRepository,
	userRepo model.IUserRepository,
	ticketHistoryRepo model.ITicketHistoryRepository,
	hub *ws.Hub,
) *TicketEngineerResolutionUsecase {
	return &TicketEngineerResolutionUsecase{
		db:                db,
		resolutionRepo:    resolutionRepo,
		ticketRepo:        ticketRepo,
		userRepo:          userRepo,
		ticketHistoryRepo: ticketHistoryRepo,
		hub:               hub,
	}
}

func (u *TicketEngineerResolutionUsecase) SubmitResolution(ctx context.Context, ticketID int64, engineerID int64, solution string, files []*multipart.FileHeader) (*model.TicketEngineerResolution, error) {
	solution = strings.TrimSpace(solution)

	if solution == "" {
		return nil, errors.New("solution wajib diisi")
	}

	engineer, err := u.userRepo.FindByID(ctx, engineerID)
	if err != nil {
		return nil, err
	}

	if engineer == nil {
		return nil, errors.New("engineer tidak ditemukan")
	}

	if engineer.RoleID != 5 {
		return nil, errors.New(
			"hanya engineer yang dapat membuat resolution",
		)
	}

	ticket, err := u.ticketRepo.FindByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if ticket == nil {
		return nil, errors.New("ticket tidak ditemukan")
	}

	if ticket.AssignedToID == nil ||
		*ticket.AssignedToID != engineerID {

		return nil, errors.New(
			"ticket bukan merupakan ticket yang sedang ditangani engineer ini",
		)
	}

	for _, file := range files {
		if file == nil {
			continue
		}

		if file.Size > MaxEngineerResolutionFileSize {
			return nil, fmt.Errorf(
				"file %s melebihi ukuran maksimal 10 MB",
				file.Filename,
			)
		}
	}

	resolution := &model.TicketEngineerResolution{
		TicketID:   ticketID,
		EngineerID: engineerID,
		Solution:   solution,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := u.resolutionRepo.Create(
		ctx,
		resolution,
	); err != nil {
		return nil, err
	}

	for _, file := range files {
		if file == nil {
			continue
		}

		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf(
				"gagal membuka file %s: %w",
				file.Filename,
				err,
			)
		}

		folder := fmt.Sprintf(
			"tickets/engineer-resolution/%d",
			ticketID,
		)

		originalName := strings.TrimSuffix(
			file.Filename,
			filepath.Ext(file.Filename),
		)

		fileName := fmt.Sprintf(
			"resolution_%d_%d_%s",
			resolution.ID,
			time.Now().UnixNano(),
			originalName,
		)

		uploadResult, err := helper.UploadFile(
			src,
			folder,
			fileName,
		)

		src.Close()

		if err != nil {
			return nil, fmt.Errorf(
				"gagal upload file %s: %w",
				file.Filename,
				err,
			)
		}

		attachment := &model.TicketEngineerResolutionAttachment{
			ResolutionID: resolution.ID,
			FileName:     file.Filename,
			FileURL:      uploadResult.SecureURL,
			PublicID:     &uploadResult.PublicID,
			CreatedAt:    time.Now(),
		}

		if err := u.resolutionRepo.CreateAttachment(
			ctx,
			attachment,
		); err != nil {
			return nil, fmt.Errorf(
				"gagal menyimpan attachment %s: %w",
				file.Filename,
				err,
			)
		}

		resolution.Attachments = append(
			resolution.Attachments,
			*attachment,
		)
	}

	history, err := u.ticketHistoryRepo.Create(
		ctx,
		model.TicketHistory{
			TicketID:  ticketID,
			UserID:    engineerID,
			Action:    "ENGINEER_RESOLUTION",
			FieldName: "engineer_resolution",
			NewValue:  &solution,
		},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"gagal menyimpan history engineer resolution: %w",
			err,
		)
	}

	histories, err := u.ticketHistoryRepo.FindByTicketID(
		ctx,
		history.TicketID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"gagal mengambil history engineer resolution: %w",
			err,
		)
	}

	if len(histories) > 0 {
		latest := histories[0]
		latest.Type = "ENGINEER_RESOLUTION"

		BroadcastTicketHistory(
			u.hub,
			latest,
		)
	}

	return resolution, nil
}

func (u *TicketEngineerResolutionUsecase) GetByTicketID(ctx context.Context, ticketID int64) (*model.TicketEngineerResolution, error) {
	return u.resolutionRepo.FindByTicketID(
		ctx,
		ticketID,
	)
}
