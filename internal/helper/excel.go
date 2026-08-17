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

	imageMaxWidth  = 210.0
	imageMaxHeight = 130.0
	rowHeight      = 110.0
)

var ticketHeaders = []string{
	"Ticket Code",
	"Project",
	"Location",
	"Asset",
	"Reporter",
	"Assigned",
	"Priority",
	"Status",
	"Foto Resolution",
	"Created At",
	"Due At",
}

var ticketColumnWidths = map[string]float64{
	"A": 20,
	"B": 20,
	"C": 20,
	"D": 20,
	"E": 20,
	"F": 20,
	"G": 15,
	"H": 15,
	"I": 30,
	"J": 20,
	"K": 20,
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
		return nil, fmt.Errorf("failed to download image: %w", err)
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
	img, _, err := image.Decode(bytes.NewReader(data))
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

	if err := png.Encode(&output, img); err != nil {
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
	if originalWidth <= 0 || originalHeight <= 0 {
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
	imageBytes, err := downloadImageFromURL(imageURL)
	if err != nil {
		return fmt.Errorf("download image: %w", err)
	}

	originalWidth, originalHeight, err := getImageDimensions(imageBytes)
	if err != nil {
		return fmt.Errorf(
			"get image dimensions: %w",
			err,
		)
	}

	pngBytes, err := convertImageToPNG(imageBytes)
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
				AltText:         "Foto Resolution",
				ScaleX:          scale,
				ScaleY:          scale,
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
	return f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: createBorders(),
	})
}

func createBorderStyle(f *excelize.File) (int, error) {
	return f.NewStyle(&excelize.Style{
		Border: createBorders(),
		Alignment: &excelize.Alignment{
			Vertical: "center",
		},
	})
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
	styles := make(map[string]int, len(statusColors))

	for status, color := range statusColors {

		style, err := f.NewStyle(&excelize.Style{
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{color},
				Pattern: 1,
			},
			Border: createBorders(),
			Alignment: &excelize.Alignment{
				Horizontal: "center",
				Vertical:   "center",
			},
		})

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

		cell, err := excelize.CoordinatesToCellName(
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
		ticket.Status,
	}

	for index, value := range values {

		cell, err := excelize.CoordinatesToCellName(
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

		if index == 7 {
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

	imageCell := fmt.Sprintf("I%d", row)

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

	createdAtCell := fmt.Sprintf("J%d", row)

	if err := f.SetCellValue(
		sheet,
		createdAtCell,
		ticket.CreatedAt.Format("2006-01-02 15:04"),
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

	dueAtCell := fmt.Sprintf("K%d", row)

	if err := f.SetCellValue(
		sheet,
		dueAtCell,
		ticket.DueAt.Format("2006-01-02 15:04"),
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

	if _, err := f.NewSheet(sheet); err != nil {
		return fmt.Errorf(
			"create summary sheet: %w",
			err,
		)
	}

	statusCount := make(map[string]int)

	for _, ticket := range tickets {
		statusCount[ticket.Status]++
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

	row := 2

	for status, count := range statusCount {

		if err := f.SetCellValue(
			sheet,
			fmt.Sprintf("A%d", row),
			status,
		); err != nil {
			return fmt.Errorf(
				"set summary status: %w",
				err,
			)
		}

		if err := f.SetCellValue(
			sheet,
			fmt.Sprintf("B%d", row),
			count,
		); err != nil {
			return fmt.Errorf(
				"set summary total: %w",
				err,
			)
		}

		row++
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

	headerStyle, err := createHeaderStyle(f)
	if err != nil {
		return nil, err
	}

	borderStyle, err := createBorderStyle(f)
	if err != nil {
		return nil, err
	}

	statusStyles, err := createStatusStyles(f)
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

	for index, ticket := range tickets {

		if ticket == nil {
			continue
		}

		row := index + 2

		if err := writeTicketRow(
			f,
			sheet,
			row,
			ticket,
			borderStyle,
			statusStyles,
		); err != nil {
			return nil, fmt.Errorf(
				"write ticket row %d: %w",
				row,
				err,
			)
		}
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
