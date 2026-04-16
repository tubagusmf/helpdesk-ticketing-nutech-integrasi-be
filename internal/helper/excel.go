package helper

import (
	"bytes"
	"strconv"

	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
	"github.com/xuri/excelize/v2"
)

func GenerateExcelTickets(tickets []*model.TicketResponse) (*bytes.Buffer, error) {
	f := excelize.NewFile()

	// ========================
	// SHEET 1: TICKETS
	// ========================
	sheet := "Tickets"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{
		"Ticket Code",
		"Project",
		"Location",
		"Asset",
		"Reporter",
		"Assigned",
		"Priority",
		"Status",
		"Created At",
		"Due At",
	}

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "FFFFFF"},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	borderStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	statusStyles := map[string]int{}

	statusColors := map[string]string{
		"OPEN":        "FFCCCC",
		"IN_PROGRESS": "FFE699",
		"RESOLVED":    "C6EFCE",
		"CLOSED":      "D9D9D9",
		"ONHOLD":      "BDD7EE",
	}

	for status, color := range statusColors {
		style, _ := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{color},
				Pattern: 1,
			},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
		})
		statusStyles[status] = style
	}

	// HEADER
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// DATA
	for i, t := range tickets {
		row := i + 2

		values := []interface{}{
			t.TicketCode,
			t.ProjectName,
			t.LocationName,
			t.AssetCode,
			t.ReporterName,
			t.AssignedToName,
			t.Priority,
			t.Status,
			t.CreatedAt.Format("2006-01-02 15:04"),
			t.DueAt.Format("2006-01-02 15:04"),
		}

		for j, val := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, val)

			if j == 7 {
				if style, ok := statusStyles[t.Status]; ok {
					f.SetCellStyle(sheet, cell, cell, style)
				}
			} else {
				f.SetCellStyle(sheet, cell, cell, borderStyle)
			}
		}
	}

	for i := 1; i <= len(headers); i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, col, col, 20)
	}

	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		Split:       false,
		XSplit:      0,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
	})

	// ========================
	// SHEET 2: SUMMARY
	// ========================
	summarySheet := "Summary"
	f.NewSheet(summarySheet)

	statusCount := map[string]int{}

	for _, t := range tickets {
		statusCount[t.Status]++
	}

	f.SetCellValue(summarySheet, "A1", "Status")
	f.SetCellValue(summarySheet, "B1", "Total")

	row := 2
	for status, count := range statusCount {
		f.SetCellValue(summarySheet, "A"+strconv.Itoa(row), status)
		f.SetCellValue(summarySheet, "B"+strconv.Itoa(row), count)
		row++
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}

	return buf, nil
}
