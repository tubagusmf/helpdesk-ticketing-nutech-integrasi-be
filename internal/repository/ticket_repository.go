package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"gorm.io/gorm"
)

type TicketRepo struct {
	db                *gorm.DB
	ticketCommentRepo model.ITicketCommentRepository
}

func NewTicketRepo(db *gorm.DB, ticketCommentRepo model.ITicketCommentRepository) model.ITicketRepository {
	return &TicketRepo{
		db:                db,
		ticketCommentRepo: ticketCommentRepo,
	}
}

func (r *TicketRepo) Create(ctx context.Context, ticket model.Ticket) (*model.Ticket, error) {
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	ticket.Status = model.StatusOpen

	if err := r.db.WithContext(ctx).Create(&ticket).Error; err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepo) FindByID(ctx context.Context, id int64) (*model.Ticket, error) {
	var ticket model.Ticket

	err := r.db.WithContext(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&ticket).Error

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepo) FindAll(
	ctx context.Context,
	filter model.Ticket,
	search string,
	startDate string,
	endDate string,
	page int,
	limit int,
	role string,
	userID int64,
) ([]*model.TicketResponse, int64, error) {

	var tickets []*model.TicketResponse
	var total int64

	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).
		Table("tickets").
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
		Joins(`
			LEFT JOIN users AS reporter
				ON reporter.id = tickets.reporter_id
		`).
		Joins(`
			LEFT JOIN users AS assigned
				ON assigned.id = tickets.assigned_to_id
		`).
		Joins(`
			LEFT JOIN ticket_resolutions
				ON ticket_resolutions.ticket_id = tickets.id
		`).
		Where(`
			tickets.deleted_at IS NULL
		`)

	if search != "" {
		s := "%" + search + "%"

		query = query.Where(`
			(
				tickets.ticket_code ILIKE ?
				OR tickets.description ILIKE ?
				OR reporter.name ILIKE ?
				OR assigned.name ILIKE ?
				OR projects.name ILIKE ?
			)
		`,
			s,
			s,
			s,
			s,
			s,
		)
	}

	if filter.TicketCode != "" {
		query = query.Where(
			"tickets.ticket_code = ?",
			filter.TicketCode,
		)
	}

	if filter.ProjectID != 0 {
		query = query.Where(
			"tickets.project_id = ?",
			filter.ProjectID,
		)
	}

	if filter.AssignedToID != nil {
		query = query.Where(
			"tickets.assigned_to_id = ?",
			*filter.AssignedToID,
		)
	}

	if filter.ReporterID != 0 {
		query = query.Where(
			"tickets.reporter_id = ?",
			filter.ReporterID,
		)
	}

	if filter.Priority != "" {
		query = query.Where(
			"tickets.priority = ?",
			filter.Priority,
		)
	}

	if filter.Status != "" {
		if role == "ENGINEER" {
			query = query.Where(`
				EXISTS (
					SELECT 1
					FROM ticket_reassignments tr
					WHERE tr.ticket_id = tickets.id
					AND tr.to_user_id = ?
					AND tr.status = ?
				)
			`, userID, filter.Status)
		} else {
			query = query.Where(
				"tickets.status = ?",
				filter.Status,
			)
		}
	}

	if startDate != "" {
		query = query.Where(
			"DATE(tickets.created_at) >= ?",
			startDate,
		)
	}

	if endDate != "" {
		query = query.Where(
			"DATE(tickets.created_at) <= ?",
			endDate,
		)
	}

	switch role {

	case "STAFF":
		query = query.Where(`
			(
				tickets.assigned_to_id = ?
				OR EXISTS (
					SELECT 1
					FROM ticket_reassignments tr
					WHERE tr.ticket_id = tickets.id
					AND tr.from_user_id = ?
				)
			)
		`, userID, userID)

	case "USER":

		query = query.
			Where(
				"tickets.reporter_id = ?",
				userID,
			).
			Where(`
				EXISTS (
					SELECT 1
					FROM user_projects up
					WHERE up.user_id = ?
					AND up.project_id = tickets.project_id
				)
			`,
				userID,
			)

	case "EXECUTIVE":

		query = query.Where(`
			EXISTS (
				SELECT 1
				FROM user_projects up
				WHERE up.user_id = ?
				AND up.project_id = tickets.project_id
			)
		`,
			userID,
		)

	case "ENGINEER":
		query = query.Where(`
			(
				tickets.assigned_to_id = ?
				OR EXISTS (
					SELECT 1
					FROM ticket_reassignments tr
					WHERE tr.ticket_id = tickets.id
					AND tr.to_user_id = ?
				)
			)
		`, userID, userID)

	case "ADMINISTRATOR":

	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	unreadQuery := `0 AS unread_comment_count`

	if role == "USER" {

		unreadQuery = `
			(
				SELECT COUNT(*)
				FROM ticket_comments tc
				WHERE tc.ticket_id = tickets.id
				AND tc.user_id != ` + fmt.Sprint(userID) + `
				AND tc.is_read_by_user = false
			) AS unread_comment_count
		`

	} else if role == "STAFF" || role == "ENGINEER" {

		unreadQuery = `
			(
				SELECT COUNT(*)
				FROM ticket_comments tc
				WHERE tc.ticket_id = tickets.id
				AND tc.user_id != ` + fmt.Sprint(userID) + `
				AND tc.is_read_by_staff = false
			) AS unread_comment_count
		`

	} else if role == "ADMINISTRATOR" {

		unreadQuery = `
			(
				SELECT COUNT(*)
				FROM ticket_comments tc
				WHERE tc.ticket_id = tickets.id
				AND tc.user_id != ` + fmt.Sprint(userID) + `
				AND tc.is_read_by_administrator = false
			) AS unread_comment_count
		`
	}

	reassignedSelect := `
    NULL AS reassigned_at
`

	engineerStatusSelect := `
    '' AS engineer_status
`

	var selectArgs []interface{}

	if role == "ENGINEER" {

		reassignedSelect = `
        (
            SELECT MAX(tr.created_at)
            FROM ticket_reassignments tr
            WHERE tr.ticket_id = tickets.id
            AND tr.to_user_id = ?
        ) AS reassigned_at
    `

		engineerStatusSelect = `
        COALESCE(
            (
                SELECT tr.status
                FROM ticket_reassignments tr
                WHERE tr.ticket_id = tickets.id
                AND tr.to_user_id = ?
                ORDER BY tr.created_at DESC, tr.id DESC
                LIMIT 1
            ),
            ''
        ) AS engineer_status
    `

		selectArgs = append(
			selectArgs,
			userID,
			userID,
		)
	}

	selectQuery := `
		tickets.id,
		tickets.ticket_code,
		tickets.project_id,
		tickets.priority,
		tickets.status,
		tickets.description,
		tickets.onhold_notes,
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

		ticket_resolutions.attachment_url AS solution_attachment,

		projects.name AS project_name,
		locations.name AS location_name,
		parts.name AS part_name,
		asset_ids.name AS asset_code,

		reporter.name AS reporter_name,
		assigned.name AS assigned_to_name,

		` + reassignedSelect + `,
		` + engineerStatusSelect + `,
		` + unreadQuery

	if role == "ENGINEER" {

		engineerUserID := fmt.Sprint(userID)

		query = query.
			Order(`
					(
						SELECT MAX(tr.created_at)
						FROM ticket_reassignments tr
						WHERE tr.ticket_id = tickets.id
						AND tr.to_user_id = ` + engineerUserID + `
					) DESC NULLS LAST
				`).
			Order(
				"tickets.created_at DESC",
			)

	} else {

		query = query.Order(
			"tickets.created_at DESC",
		)
	}

	if err := query.
		Select(
			selectQuery,
			selectArgs...,
		).
		Limit(limit).
		Offset(offset).
		Scan(&tickets).
		Error; err != nil {

		return nil, 0, err
	}

	for _, ticket := range tickets {

		count, err := r.ticketCommentRepo.CountUnreadByTicket(
			ctx,
			ticket.ID,
			role,
			userID,
		)

		if err != nil {
			return nil, 0, err
		}

		ticket.UnreadCommentCount = count
	}

	return tickets, total, nil
}

func (r *TicketRepo) Update(ctx context.Context, ticket model.Ticket) error {
	ticket.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ? AND deleted_at IS NULL", ticket.ID).
		Updates(ticket).Error
}

func (r *TicketRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error
}

func (r *TicketRepo) CountByProjectToday(ctx context.Context, projectID int64) (int64, error) {
	var count int64

	err := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("project_id = ?", projectID).
		Where("DATE(created_at) = CURRENT_DATE").
		Count(&count).Error

	return count, err
}

func (r *TicketRepo) FindResponseByID(ctx context.Context, id int64) (*model.TicketResponse, error) {
	var ticket model.TicketResponse

	err := r.db.WithContext(ctx).
		Table("tickets").
		Select(`
        tickets.id,
        tickets.ticket_code,
        tickets.project_id,
        tickets.priority,
        tickets.status,
        tickets.description,
        tickets.onhold_notes,
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

        projects.name as project_name,
        locations.name as location_name,
        parts.name as part_name,
        asset_ids.name as asset_code,
        reporter.name as reporter_name,
        assigned.name as assigned_to_name,

        (
            SELECT tr.status
            FROM ticket_reassignments tr
            WHERE tr.ticket_id = tickets.id
            ORDER BY tr.created_at DESC, tr.id DESC
            LIMIT 1
        ) AS engineer_status
    `).
		Joins("LEFT JOIN projects ON projects.id = tickets.project_id").
		Joins("LEFT JOIN locations ON locations.id = tickets.location_id").
		Joins("LEFT JOIN parts ON parts.id = tickets.part_id").
		Joins("LEFT JOIN asset_ids ON asset_ids.id = tickets.asset_id").
		Joins("LEFT JOIN users as reporter ON reporter.id = tickets.reporter_id").
		Joins("LEFT JOIN users as assigned ON assigned.id = tickets.assigned_to_id").
		Where("tickets.id = ?", id).
		Scan(&ticket).Error

	if err != nil {
		return nil, err
	}

	return &ticket, nil
}

func (r *TicketRepo) SetStaffAssignment(ctx context.Context, ticketID int64, staffID int64) error {
	var roleID int

	err := r.db.WithContext(ctx).
		Table("users").
		Select("role_id").
		Where("id = ?", staffID).
		Scan(&roleID).Error

	if err != nil {
		return err
	}

	if roleID != 2 {
		return fmt.Errorf("assigned user must be STAFF")
	}

	now := time.Now()

	return r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Updates(map[string]interface{}{
			"staff_assigned_to_id":        staffID,
			"staff_assigned_at":           now,
			"staff_first_response_at":     nil,
			"staff_response_time_seconds": nil,
			"updated_at":                  now,
		}).Error
}

func (r *TicketRepo) RecordStaffFirstResponse(ctx context.Context, ticketID int64, userID int64) error {
	var ticket model.Ticket

	if err := r.db.WithContext(ctx).
		Where("id = ?", ticketID).
		First(&ticket).Error; err != nil {
		return err
	}

	if ticket.StaffAssignedToID == nil ||
		ticket.StaffAssignedAt == nil {
		return nil
	}

	if ticket.StaffFirstResponseAt != nil {
		return nil
	}

	var roleID int

	if err := r.db.WithContext(ctx).
		Table("users").
		Select("role_id").
		Where("id = ?", userID).
		Scan(&roleID).Error; err != nil {
		return err
	}

	if roleID != 2 {
		return nil
	}

	if *ticket.StaffAssignedToID != userID {
		return nil
	}

	now := time.Now()

	responseSeconds := int64(
		now.Sub(*ticket.StaffAssignedAt).Seconds(),
	)

	if responseSeconds < 0 {
		responseSeconds = 0
	}

	result := r.db.WithContext(ctx).
		Model(&model.Ticket{}).
		Where("id = ?", ticketID).
		Where("staff_assigned_to_id = ?", userID).
		Where("staff_first_response_at IS NULL").
		Updates(map[string]interface{}{
			"staff_first_response_at":     now,
			"staff_response_time_seconds": responseSeconds,
			"updated_at":                  now,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}
