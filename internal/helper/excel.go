package helper

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"time"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"github.com/xuri/excelize/v2"
	_ "golang.org/x/image/webp"
)

const (
	ticketSheetName  = "Tickets"
	summarySheetName = "Summary"
	imageMaxWidth    = 140.0
	imageMaxHeight   = 80.0
	rowHeight        = 82.0
	headerRowHeight  = 24.0
	customRowHeight  = 24.0
)

type CustomExcelColumn struct {
	Key   string
	Title string
	Width float64
}

var customTicketColumns = map[string]CustomExcelColumn{

	"ticket_code": {
		Key:   "ticket_code",
		Title: "Ticket Code",
		Width: 18,
	},

	"project": {
		Key:   "project",
		Title: "Project",
		Width: 18,
	},

	"location": {
		Key:   "location",
		Title: "Location",
		Width: 17,
	},

	"part": {
		Key:   "part",
		Title: "Part",
		Width: 17,
	},

	"asset": {
		Key:   "asset",
		Title: "Asset",
		Width: 15,
	},

	"reporter": {
		Key:   "reporter",
		Title: "Reporter",
		Width: 18,
	},

	"assigned_engineer": {
		Key:   "assigned_engineer",
		Title: "Assigned Engineer",
		Width: 19,
	},

	"staff_assigned": {
		Key:   "staff_assigned",
		Title: "Staff Assigned",
		Width: 18,
	},

	"priority": {
		Key:   "priority",
		Title: "Priority",
		Width: 12,
	},

	"status": {
		Key:   "status",
		Title: "Status",
		Width: 14,
	},

	"description": {
		Key:   "description",
		Title: "Permasalahan",
		Width: 28,
	},

	"onhold_notes": {
		Key:   "onhold_notes",
		Title: "Onhold Notes",
		Width: 25,
	},

	"attachment": {
		Key:   "attachment",
		Title: "Ticket Attachment",
		Width: 24,
	},

	"resolution_cause": {
		Key:   "resolution_cause",
		Title: "Resolution Cause",
		Width: 24,
	},

	"resolution_solution": {
		Key:   "resolution_solution",
		Title: "Resolution Solution",
		Width: 26,
	},

	"resolution_notes": {
		Key:   "resolution_notes",
		Title: "Resolution Notes",
		Width: 28,
	},

	"resolution_completion_at": {
		Key:   "resolution_completion_at",
		Title: "Resolution Completion Time",
		Width: 22,
	},

	"resolution_attachment": {
		Key:   "resolution_attachment",
		Title: "Resolution Attachment",
		Width: 24,
	},

	"created_at": {
		Key:   "created_at",
		Title: "Created At",
		Width: 18,
	},

	"updated_at": {
		Key:   "updated_at",
		Title: "Updated At",
		Width: 18,
	},

	"due_at": {
		Key:   "due_at",
		Title: "Due At",
		Width: 18,
	},

	"resolved_at": {
		Key:   "resolved_at",
		Title: "Resolved At",
		Width: 18,
	},

	"staff_assigned_at": {
		Key:   "staff_assigned_at",
		Title: "Staff Assigned At",
		Width: 20,
	},

	"staff_first_response_at": {
		Key:   "staff_first_response_at",
		Title: "Staff First Response At",
		Width: 22,
	},

	"staff_response_time": {
		Key:   "staff_response_time",
		Title: "Staff Response Time",
		Width: 20,
	},

	"engineer_resolution_at": {
		Key:   "engineer_resolution_at",
		Title: "Engineer Resolution At",
		Width: 22,
	},
}

var ticketHeaders = []string{
	"Ticket Code",
	"Project",
	"Location",
	"Asset",
	"Reporter",
	"Assigned",
	"Priority",
	"Permasalahan",
	"Penyebab",
	"Solusi",
	"Status",
	"Foto Sesudah",
	"Created At",
	"Due At",
}

var ticketColumnWidths = map[string]float64{
	"A": 18, // Ticket Code
	"B": 18, // Project
	"C": 17, // Location
	"D": 15, // Asset
	"E": 18, // Reporter
	"F": 18, // Assigned
	"G": 12, // Priority
	"H": 28, // Permasalahan
	"I": 22, // Penyebab
	"J": 26, // Solusi
	"K": 14, // Status
	"L": 22, // Foto Sesudah
	"M": 18, // Created At
	"N": 18, // Due At
}

var statusColors = map[string]string{
	"OPEN":        "FFCCCC",
	"IN_PROGRESS": "FFE699",
	"RESOLVED":    "C6EFCE",
	"CLOSED":      "D9D9D9",
	"ONHOLD":      "BDD7EE",
}

func downloadImageFromURL(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to download image: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"failed to download image, status: %s",
			resp.Status,
		)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read image: %w",
			err,
		)
	}

	return data, nil
}

func decodeImage(data []byte) (image.Image, error) {
	img, _, err := image.Decode(
		bytes.NewReader(data),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to decode image: %w",
			err,
		)
	}

	return img, nil
}

func convertImageToPNG(data []byte) ([]byte, error) {
	img, err := decodeImage(data)
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer

	if err := png.Encode(
		&output,
		img,
	); err != nil {
		return nil, fmt.Errorf(
			"failed to encode image to PNG: %w",
			err,
		)
	}

	return output.Bytes(), nil
}

func getImageDimensions(data []byte) (int, int, error) {
	img, err := decodeImage(data)
	if err != nil {
		return 0, 0, err
	}

	bounds := img.Bounds()

	return bounds.Dx(), bounds.Dy(), nil
}

func calculateImageScale(originalWidth int, originalHeight int, maxWidth float64, maxHeight float64) float64 {
	if originalWidth <= 0 ||
		originalHeight <= 0 {
		return 1
	}

	width := float64(originalWidth)
	height := float64(originalHeight)

	scaleX := maxWidth / width
	scaleY := maxHeight / height

	if scaleX < scaleY {
		return scaleX
	}

	return scaleY
}

func addResolutionImage(f *excelize.File, sheet string, cell string, imageURL string) error {
	imageBytes, err := downloadImageFromURL(
		imageURL,
	)

	if err != nil {
		return fmt.Errorf(
			"download image: %w",
			err,
		)
	}

	originalWidth,
		originalHeight,
		err := getImageDimensions(
		imageBytes,
	)

	if err != nil {
		return fmt.Errorf(
			"get image dimensions: %w",
			err,
		)
	}

	pngBytes, err := convertImageToPNG(
		imageBytes,
	)

	if err != nil {
		return fmt.Errorf(
			"convert image to PNG: %w",
			err,
		)
	}

	scale := calculateImageScale(
		originalWidth,
		originalHeight,
		imageMaxWidth,
		imageMaxHeight,
	)

	err = f.AddPictureFromBytes(
		sheet,
		cell,
		&excelize.Picture{
			Extension: ".png",
			File:      pngBytes,
			Format: &excelize.GraphicOptions{
				AltText: "Foto Sesudah",
				ScaleX:  scale,
				ScaleY:  scale,

				LockAspectRatio: true,
				Positioning:     "oneCell",
			},
		},
	)

	if err != nil {
		return fmt.Errorf(
			"insert image into Excel: %w",
			err,
		)
	}

	return nil
}

func createHeaderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(
		&excelize.Style{
			Font: &excelize.Font{
				Bold:  true,
				Color: "FFFFFF",
				Size:  10,
			},

			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{"4472C4"},
				Pattern: 1,
			},

			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
				WrapText:   true,
			},

			Border: createBorders(),
		},
	)
}

func createBorderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(
		&excelize.Style{
			Font: &excelize.Font{
				Size: 9,
			},

			Border: createBorders(),

			Alignment: &excelize.Alignment{
				Vertical: "center",
				WrapText: true,
			},
		},
	)
}

func createBorders() []excelize.Border {
	return []excelize.Border{
		{
			Type:  "left",
			Color: "000000",
			Style: 1,
		},
		{
			Type:  "right",
			Color: "000000",
			Style: 1,
		},
		{
			Type:  "top",
			Color: "000000",
			Style: 1,
		},
		{
			Type:  "bottom",
			Color: "000000",
			Style: 1,
		},
	}
}

func createStatusStyles(f *excelize.File) (map[string]int, error) {
	styles := make(
		map[string]int,
		len(statusColors),
	)

	for status, color := range statusColors {

		style, err := f.NewStyle(
			&excelize.Style{
				Font: &excelize.Font{
					Size: 9,
				},

				Fill: excelize.Fill{
					Type:    "pattern",
					Color:   []string{color},
					Pattern: 1,
				},

				Border: createBorders(),

				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
					WrapText:   true,
				},
			},
		)

		if err != nil {
			return nil, fmt.Errorf(
				"create status style for %s: %w",
				status,
				err,
			)
		}

		styles[status] = style
	}

	return styles, nil
}

func writeTicketHeaders(f *excelize.File, sheet string, styleID int) error {
	for i, header := range ticketHeaders {

		cell, err :=
			excelize.CoordinatesToCellName(
				i+1,
				1,
			)

		if err != nil {
			return fmt.Errorf(
				"create header cell: %w",
				err,
			)
		}

		if err := f.SetCellValue(
			sheet,
			cell,
			header,
		); err != nil {
			return fmt.Errorf(
				"set header value %s: %w",
				cell,
				err,
			)
		}

		if err := f.SetCellStyle(
			sheet,
			cell,
			cell,
			styleID,
		); err != nil {
			return fmt.Errorf(
				"set header style %s: %w",
				cell,
				err,
			)
		}
	}

	if err := f.SetRowHeight(
		sheet,
		1,
		headerRowHeight,
	); err != nil {
		return fmt.Errorf(
			"set header row height: %w",
			err,
		)
	}

	return nil
}

func writeTicketRow(f *excelize.File, sheet string, row int, ticket *model.TicketResponse, borderStyle int, statusStyles map[string]int) error {
	values := []interface{}{
		ticket.TicketCode,
		ticket.ProjectName,
		ticket.LocationName,
		ticket.AssetCode,
		ticket.ReporterName,
		ticket.AssignedToName,
		ticket.Priority,
		ticket.Description,
		ticket.CauseName,
		ticket.SolutionName,
		ticket.Status,
	}

	for index, value := range values {

		cell, err :=
			excelize.CoordinatesToCellName(
				index+1,
				row,
			)

		if err != nil {
			return fmt.Errorf(
				"create ticket cell: %w",
				err,
			)
		}

		if err := f.SetCellValue(
			sheet,
			cell,
			value,
		); err != nil {
			return fmt.Errorf(
				"set ticket value %s: %w",
				cell,
				err,
			)
		}

		styleID := borderStyle

		if index == 10 {
			if statusStyle, ok := statusStyles[ticket.Status]; ok {
				styleID = statusStyle
			}
		}

		if err := f.SetCellStyle(
			sheet,
			cell,
			cell,
			styleID,
		); err != nil {
			return fmt.Errorf(
				"set ticket style %s: %w",
				cell,
				err,
			)
		}
	}

	imageCell := fmt.Sprintf(
		"L%d",
		row,
	)

	if err := f.SetCellStyle(
		sheet,
		imageCell,
		imageCell,
		borderStyle,
	); err != nil {
		return fmt.Errorf(
			"set image cell style: %w",
			err,
		)
	}

	if ticket.SolutionAttachment != nil &&
		*ticket.SolutionAttachment != "" {

		if err := addResolutionImage(
			f,
			sheet,
			imageCell,
			*ticket.SolutionAttachment,
		); err != nil {

			fmt.Printf(
				"failed to add resolution image for ticket %s: %v\n",
				ticket.TicketCode,
				err,
			)

			if err := f.SetCellValue(
				sheet,
				imageCell,
				"Failed to insert image",
			); err != nil {
				return fmt.Errorf(
					"set image error message: %w",
					err,
				)
			}
		}
	}

	createdAtCell := fmt.Sprintf(
		"M%d",
		row,
	)

	if err := f.SetCellValue(
		sheet,
		createdAtCell,
		ticket.CreatedAt.Format(
			"2006-01-02 15:04",
		),
	); err != nil {
		return fmt.Errorf(
			"set created at: %w",
			err,
		)
	}

	if err := f.SetCellStyle(
		sheet,
		createdAtCell,
		createdAtCell,
		borderStyle,
	); err != nil {
		return fmt.Errorf(
			"set created at style: %w",
			err,
		)
	}

	dueAtCell := fmt.Sprintf(
		"N%d",
		row,
	)

	if err := f.SetCellValue(
		sheet,
		dueAtCell,
		ticket.DueAt.Format(
			"2006-01-02 15:04",
		),
	); err != nil {
		return fmt.Errorf(
			"set due at: %w",
			err,
		)
	}

	if err := f.SetCellStyle(
		sheet,
		dueAtCell,
		dueAtCell,
		borderStyle,
	); err != nil {
		return fmt.Errorf(
			"set due at style: %w",
			err,
		)
	}

	if err := f.SetRowHeight(
		sheet,
		row,
		rowHeight,
	); err != nil {
		return fmt.Errorf(
			"set row height: %w",
			err,
		)
	}

	return nil
}

func setTicketColumnWidths(f *excelize.File, sheet string) error {
	for column, width := range ticketColumnWidths {

		if err := f.SetColWidth(
			sheet,
			column,
			column,
			width,
		); err != nil {
			return fmt.Errorf(
				"set column width %s: %w",
				column,
				err,
			)
		}
	}

	return nil
}

func setTicketFreezePane(f *excelize.File, sheet string) error {
	if err := f.SetPanes(
		sheet,
		&excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		},
	); err != nil {
		return fmt.Errorf(
			"set freeze pane: %w",
			err,
		)
	}

	return nil
}

func writeSummarySheet(f *excelize.File, tickets []*model.TicketResponse) error {
	sheet := summarySheetName

	if _, err := f.NewSheet(
		sheet,
	); err != nil {
		return fmt.Errorf(
			"create summary sheet: %w",
			err,
		)
	}

	statusCount := make(
		map[string]int,
	)

	for _, ticket := range tickets {

		if ticket == nil {
			continue
		}

		statusCount[ticket.Status]++
	}

	headerStyle, err := createHeaderStyle(
		f,
	)

	if err != nil {
		return err
	}

	bodyStyle, err := createBorderStyle(
		f,
	)

	if err != nil {
		return err
	}

	if err := f.SetCellValue(
		sheet,
		"A1",
		"Status",
	); err != nil {
		return fmt.Errorf(
			"set summary status header: %w",
			err,
		)
	}

	if err := f.SetCellValue(
		sheet,
		"B1",
		"Total",
	); err != nil {
		return fmt.Errorf(
			"set summary total header: %w",
			err,
		)
	}

	if err := f.SetCellStyle(
		sheet,
		"A1",
		"B1",
		headerStyle,
	); err != nil {
		return fmt.Errorf(
			"set summary header style: %w",
			err,
		)
	}

	row := 2

	for status, count := range statusCount {

		if err := f.SetCellValue(
			sheet,
			fmt.Sprintf(
				"A%d",
				row,
			),
			status,
		); err != nil {
			return fmt.Errorf(
				"set summary status: %w",
				err,
			)
		}

		if err := f.SetCellValue(
			sheet,
			fmt.Sprintf(
				"B%d",
				row,
			),
			count,
		); err != nil {
			return fmt.Errorf(
				"set summary total: %w",
				err,
			)
		}

		if err := f.SetCellStyle(
			sheet,
			fmt.Sprintf(
				"A%d",
				row,
			),
			fmt.Sprintf(
				"B%d",
				row,
			),
			bodyStyle,
		); err != nil {
			return fmt.Errorf(
				"set summary body style: %w",
				err,
			)
		}

		row++
	}

	if err := f.SetColWidth(
		sheet,
		"A",
		"A",
		18,
	); err != nil {
		return fmt.Errorf(
			"set summary status width: %w",
			err,
		)
	}

	if err := f.SetColWidth(
		sheet,
		"B",
		"B",
		12,
	); err != nil {
		return fmt.Errorf(
			"set summary total width: %w",
			err,
		)
	}

	return nil
}

func GenerateExcelTickets(tickets []*model.TicketResponse) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	sheet := ticketSheetName

	if err := f.SetSheetName(
		"Sheet1",
		sheet,
	); err != nil {
		return nil, fmt.Errorf(
			"rename ticket sheet: %w",
			err,
		)
	}

	headerStyle, err := createHeaderStyle(
		f,
	)

	if err != nil {
		return nil, err
	}

	borderStyle, err := createBorderStyle(
		f,
	)

	if err != nil {
		return nil, err
	}

	statusStyles, err := createStatusStyles(
		f,
	)

	if err != nil {
		return nil, err
	}

	if err := writeTicketHeaders(
		f,
		sheet,
		headerStyle,
	); err != nil {
		return nil, err
	}

	excelRow := 2

	for _, ticket := range tickets {

		if ticket == nil {
			continue
		}

		if err := writeTicketRow(
			f,
			sheet,
			excelRow,
			ticket,
			borderStyle,
			statusStyles,
		); err != nil {
			return nil, fmt.Errorf(
				"write ticket row %d: %w",
				excelRow,
				err,
			)
		}

		excelRow++
	}

	if err := setTicketColumnWidths(
		f,
		sheet,
	); err != nil {
		return nil, err
	}

	if err := setTicketFreezePane(
		f,
		sheet,
	); err != nil {
		return nil, err
	}

	if excelRow > 2 {

		if err := f.AutoFilter(
			sheet,
			fmt.Sprintf(
				"A1:N%d",
				excelRow-1,
			),
			nil,
		); err != nil {
			return nil, fmt.Errorf(
				"set ticket autofilter: %w",
				err,
			)
		}
	}

	if err := writeSummarySheet(
		f,
		tickets,
	); err != nil {
		return nil, err
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf(
			"write Excel buffer: %w",
			err,
		)
	}

	return buf, nil
}

func ValidateCustomTicketExportColumns(columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf(
			"at least one column must be selected",
		)
	}

	seen := make(
		map[string]bool,
		len(columns),
	)

	for _, column := range columns {

		if _, ok :=
			customTicketColumns[column]; !ok {

			return fmt.Errorf(
				"invalid export column: %s",
				column,
			)
		}

		if seen[column] {
			return fmt.Errorf(
				"duplicate export column: %s",
				column,
			)
		}

		seen[column] = true
	}

	return nil
}

func customStringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}

func customTimeValue(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.Format(
		"2006-01-02 15:04:05",
	)
}

func formatResponseTime(seconds *int64) string {
	if seconds == nil {
		return ""
	}

	total := *seconds

	if total < 0 {
		total = 0
	}

	hours := total / 3600
	minutes := (total % 3600) / 60
	secs := total % 60

	return fmt.Sprintf(
		"%02d:%02d:%02d",
		hours,
		minutes,
		secs,
	)
}

func getCustomTicketValue(ticket *model.TicketExportRow, column string) interface{} {
	switch column {

	case "ticket_code":
		return ticket.TicketCode

	case "project":
		return ticket.ProjectName

	case "location":
		return ticket.LocationName

	case "part":
		return ticket.PartName

	case "asset":
		return ticket.AssetCode

	case "reporter":
		return ticket.ReporterName

	case "assigned_engineer":
		return ticket.AssignedToName

	case "staff_assigned":
		return ticket.StaffAssignedToName

	case "priority":
		return ticket.Priority

	case "status":
		return ticket.Status

	case "description":
		return ticket.Description

	case "onhold_notes":
		return customStringValue(
			ticket.OnholdNotes,
		)

	case "attachment":
		return customStringValue(
			ticket.TicketAttachmentURL,
		)

	case "resolution_cause":
		return customStringValue(
			ticket.ResolutionCause,
		)

	case "resolution_solution":
		return customStringValue(
			ticket.ResolutionSolution,
		)

	case "resolution_notes":
		return customStringValue(
			ticket.ResolutionNotes,
		)

	case "resolution_completion_at":
		return customTimeValue(
			ticket.ResolutionCompletionAt,
		)

	case "resolution_attachment":
		return customStringValue(
			ticket.ResolutionAttachmentURL,
		)

	case "created_at":
		return ticket.CreatedAt.Format(
			"2006-01-02 15:04:05",
		)

	case "updated_at":
		return ticket.UpdatedAt.Format(
			"2006-01-02 15:04:05",
		)

	case "due_at":
		return ticket.DueAt.Format(
			"2006-01-02 15:04:05",
		)

	case "resolved_at":
		return customTimeValue(
			ticket.ResolvedAt,
		)

	case "staff_assigned_at":
		return customTimeValue(
			ticket.StaffAssignedAt,
		)

	case "staff_first_response_at":
		return customTimeValue(
			ticket.StaffFirstResponseAt,
		)

	case "staff_response_time":
		return formatResponseTime(
			ticket.StaffResponseTimeSeconds,
		)

	case "engineer_resolution_at":
		return customTimeValue(
			ticket.EngineerResolutionAt,
		)

	default:
		return ""
	}
}

func GenerateCustomExcelTickets(tickets []*model.TicketExportRow, columns []string) (*bytes.Buffer, error) {
	if err := ValidateCustomTicketExportColumns(
		columns,
	); err != nil {
		return nil, err
	}

	f := excelize.NewFile()

	sheet := ticketSheetName

	if err := f.SetSheetName(
		"Sheet1",
		sheet,
	); err != nil {
		return nil, fmt.Errorf(
			"rename sheet: %w",
			err,
		)
	}

	headerStyle, err := createHeaderStyle(
		f,
	)

	if err != nil {
		return nil, err
	}

	borderStyle, err := createBorderStyle(
		f,
	)

	if err != nil {
		return nil, err
	}

	statusStyles, err := createStatusStyles(
		f,
	)

	if err != nil {
		return nil, err
	}

	for index, column := range columns {

		columnConfig :=
			customTicketColumns[column]

		cell, err :=
			excelize.CoordinatesToCellName(
				index+1,
				1,
			)

		if err != nil {
			return nil, fmt.Errorf(
				"create custom header cell: %w",
				err,
			)
		}

		if err := f.SetCellValue(
			sheet,
			cell,
			columnConfig.Title,
		); err != nil {
			return nil, fmt.Errorf(
				"set custom header %s: %w",
				cell,
				err,
			)
		}

		if err := f.SetCellStyle(
			sheet,
			cell,
			cell,
			headerStyle,
		); err != nil {
			return nil, fmt.Errorf(
				"set custom header style %s: %w",
				cell,
				err,
			)
		}

		columnName, err :=
			excelize.ColumnNumberToName(
				index + 1,
			)

		if err != nil {
			return nil, fmt.Errorf(
				"get custom column name: %w",
				err,
			)
		}

		if err := f.SetColWidth(
			sheet,
			columnName,
			columnName,
			columnConfig.Width,
		); err != nil {
			return nil, fmt.Errorf(
				"set custom column width %s: %w",
				columnName,
				err,
			)
		}
	}

	if err := f.SetRowHeight(
		sheet,
		1,
		headerRowHeight,
	); err != nil {
		return nil, fmt.Errorf(
			"set custom header row height: %w",
			err,
		)
	}

	rowNumber := 2

	for _, ticket := range tickets {

		if ticket == nil {
			continue
		}

		for index, column := range columns {

			cell, err :=
				excelize.CoordinatesToCellName(
					index+1,
					rowNumber,
				)

			if err != nil {
				return nil, fmt.Errorf(
					"create custom cell: %w",
					err,
				)
			}

			value :=
				getCustomTicketValue(
					ticket,
					column,
				)

			if err := f.SetCellValue(
				sheet,
				cell,
				value,
			); err != nil {
				return nil, fmt.Errorf(
					"set custom value %s: %w",
					cell,
					err,
				)
			}

			styleID := borderStyle

			if column == "status" {

				if statusStyle, ok :=
					statusStyles[ticket.Status]; ok {

					styleID = statusStyle
				}
			}

			if err := f.SetCellStyle(
				sheet,
				cell,
				cell,
				styleID,
			); err != nil {
				return nil, fmt.Errorf(
					"set custom style %s: %w",
					cell,
					err,
				)
			}
		}

		if err := f.SetRowHeight(
			sheet,
			rowNumber,
			customRowHeight,
		); err != nil {
			return nil, fmt.Errorf(
				"set custom row height: %w",
				err,
			)
		}

		rowNumber++
	}

	if err := f.SetPanes(
		sheet,
		&excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      0,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		},
	); err != nil {
		return nil, fmt.Errorf(
			"set custom freeze pane: %w",
			err,
		)
	}

	if rowNumber > 2 {

		lastColumnName, err :=
			excelize.ColumnNumberToName(
				len(columns),
			)

		if err != nil {
			return nil, fmt.Errorf(
				"get custom last column: %w",
				err,
			)
		}

		if err := f.AutoFilter(
			sheet,
			fmt.Sprintf(
				"A1:%s%d",
				lastColumnName,
				rowNumber-1,
			),
			nil,
		); err != nil {
			return nil, fmt.Errorf(
				"set custom autofilter: %w",
				err,
			)
		}
	}

	buf, err := f.WriteToBuffer()

	if err != nil {
		return nil, fmt.Errorf(
			"write custom Excel buffer: %w",
			err,
		)
	}

	return buf, nil
}
