package helper

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/phpdave11/gofpdf"
	"github.com/tubagusmf/helpdesk-ticketing-nutech-integrasi-be/internal/model"
)

func GeneratePDFTicketHistory(ticketCode string, ticketID int64, histories []*model.TicketHistoryResponse) (*bytes.Buffer, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 18, 15)
	pdf.SetAutoPageBreak(true, 17)
	pdf.SetTitle("Riwayat Tiket "+ticketCode, true)
	pdf.SetAuthor("Helpdesk Ticketing", true)
	pdf.AliasNbPages("{nb}")

	pdf.SetHeaderFunc(func() {
		pdf.SetY(7)
		pdf.SetFont("Arial", "B", 8)
		pdf.SetTextColor(100, 116, 139)

		pdf.CellFormat(
			0, 5,
			"HELPDESK TICKETING | RIWAYAT AKTIVITAS",
			"", 1, "L", false, 0, "",
		)

		pdf.SetDrawColor(226, 232, 240)
		pdf.Line(15, 14, 195, 14)
		pdf.SetY(18)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-12)
		pdf.SetFont("Arial", "I", 8)
		pdf.SetTextColor(107, 114, 128)

		pdf.CellFormat(
			0, 6,
			fmt.Sprintf("Halaman %d dari {nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "",
		)
	})

	pdf.AddPage()

	pdf.SetFont("Arial", "B", 17)
	pdf.SetTextColor(30, 41, 59)
	pdf.CellFormat(
		0, 9,
		"RIWAYAT AKTIVITAS TIKET",
		"", 1, "L", false, 0, "",
	)

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(51, 65, 85)
	pdf.CellFormat(
		0, 7,
		"Ticket: "+ticketCode,
		"", 1, "L", false, 0, "",
	)

	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(100, 116, 139)

	pdf.CellFormat(
		0, 6,
		fmt.Sprintf(
			"ID Tiket: %d | Total aktivitas: %d",
			ticketID,
			len(histories),
		),
		"", 1, "L", false, 0, "",
	)

	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		location = time.Local
	}

	pdf.CellFormat(
		0, 5,
		"Dicetak: "+time.Now().In(location).Format("02/01/2006 15:04:05"),
		"", 1, "L", false, 0, "",
	)

	pdf.Ln(3)

	if len(histories) == 0 {
		pdf.SetFont("Arial", "I", 10)
		pdf.SetTextColor(100, 116, 139)
		pdf.MultiCell(
			0, 6,
			"Belum ada riwayat aktivitas untuk tiket ini.",
			"", "L", false,
		)
	}

	const (
		leftMargin = 15.0
		cardWidth  = 180.0
		innerWidth = 172.0
		lineHeight = 4.5
	)

	for _, history := range histories {
		if history == nil {
			continue
		}

		label := ticketHistoryPDFTitle(history.Type)
		detail := ticketHistoryPDFDetail(history)

		actor := strings.TrimSpace(history.UserName)
		if ticketHistoryPDFIsSystem(history.Type) {
			actor = "System"
		} else if actor == "" {
			actor = "-"
		}

		createdAt := history.CreatedAt.In(location).
			Format("02/01/2006, 15:04:05")

		pdf.SetFont("Arial", "B", 9)
		labelLines := pdf.SplitLines([]byte(label), innerWidth)

		pdf.SetFont("Arial", "", 9)
		bodyLines := pdf.SplitLines([]byte(detail), innerWidth)

		if len(labelLines) == 0 {
			labelLines = [][]byte{[]byte(label)}
		}

		if len(bodyLines) == 0 {
			bodyLines = [][]byte{[]byte(detail)}
		}

		labelHeight := float64(len(labelLines)) * lineHeight
		bodyHeight := float64(len(bodyLines)) * lineHeight

		cardHeight := 15 + labelHeight + bodyHeight

		if pdf.GetY()+cardHeight > 279 {
			pdf.AddPage()
		}

		x := leftMargin
		y := pdf.GetY()

		pdf.SetDrawColor(220, 226, 234)
		pdf.SetFillColor(255, 255, 255)
		pdf.Rect(x, y, cardWidth, cardHeight, "FD")
		pdf.SetXY(x+4, y+3)
		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(100, 116, 139)

		pdf.CellFormat(
			92, 5, actor,
			"", 0, "L", false, 0, "",
		)

		pdf.CellFormat(
			innerWidth-92, 5, createdAt,
			"", 1, "R", false, 0, "",
		)

		labelY := y + 10

		pdf.SetXY(x+4, labelY)
		pdf.SetFont("Arial", "B", 9)
		pdf.SetTextColor(30, 41, 59)
		pdf.MultiCell(
			innerWidth, lineHeight,
			label, "", "L", false,
		)

		detailY := labelY + labelHeight + 2

		pdf.SetXY(x+4, detailY)
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(51, 65, 85)
		pdf.MultiCell(
			innerWidth, lineHeight,
			detail, "", "L", false,
		)

		pdf.SetY(y + cardHeight + 3)
		pdf.SetX(leftMargin)
	}

	var buffer bytes.Buffer

	if err := pdf.Output(&buffer); err != nil {
		return nil, err
	}

	return &buffer, nil
}

func ticketHistoryPDFValue(value *string) string {
	if value == nil {
		return "-"
	}

	result := strings.TrimSpace(*value)
	if result == "" {
		return "-"
	}

	return result
}

func ticketHistoryPDFFirst(values ...*string) string {
	for _, value := range values {
		result := ticketHistoryPDFValue(value)
		if result != "-" {
			return result
		}
	}

	return "-"
}

func ticketHistoryPDFTitle(historyType string) string {
	switch historyType {
	case "CREATED":
		return "Ticket dibuat"
	case "STATUS_UPDATED":
		return "Status diubah"
	case "COMMENT":
		return "Komentar"
	case "ONHOLD_NOTE":
		return "Catatan OnHold"
	case "REASSIGNED":
		return "Ticket dialihkan ke Engineer"
	case "STAFF_REASSIGNED":
		return "Ticket dialihkan ke Staff CCIT"
	case "STAFF_REASSIGNMENT_REASON":
		return "Justifikasi pengalihan ticket"
	case "RESOLUTION":
		return "Solusi tiket"
	case "ENGINEER_RESOLUTION":
		return "Solusi Tim Engineer"
	case "CANCELLATION_REASON":
		return "Justifikasi pembatalan"
	case "TICKET_OVERDUE":
		return "Ticket Overdue"
	case "STAFF_RESPONSE_OVERDUE":
		return "Ticket belum direspons"
	case "STAFF_RESPONSE_LATE":
		return "Ticket terlambat direspons"
	default:
		return "Aktivitas lainnya"
	}
}

func ticketHistoryPDFIsSystem(historyType string) bool {
	switch historyType {
	case "TICKET_OVERDUE",
		"STAFF_RESPONSE_OVERDUE",
		"STAFF_RESPONSE_LATE":
		return true
	default:
		return false
	}
}

func ticketHistoryPDFDetail(h *model.TicketHistoryResponse) string {
	switch h.Type {
	case "CREATED":
		return "Tiket berhasil dibuat."

	case "STATUS_UPDATED":
		return fmt.Sprintf(
			"Status: %s -> %s",
			ticketHistoryPDFValue(h.OldValue),
			ticketHistoryPDFValue(h.NewValue),
		)

	case "COMMENT":
		return ticketHistoryPDFFirst(h.Message, h.NewValue)

	case "ONHOLD_NOTE":
		return "Catatan OnHold: " +
			ticketHistoryPDFFirst(h.Message, h.NewValue)

	case "REASSIGNED":
		return "Tiket dialihkan kepada Tim Engineer."

	case "STAFF_REASSIGNED":
		oldName := strings.TrimSpace(h.OldUserName)
		newName := strings.TrimSpace(h.NewUserName)

		if oldName == "" {
			oldName = "-"
		}
		if newName == "" {
			newName = "-"
		}

		return fmt.Sprintf(
			"Penanggung jawab Staff CCIT: %s -> %s",
			oldName,
			newName,
		)

	case "STAFF_REASSIGNMENT_REASON":
		return ticketHistoryPDFFirst(h.Message, h.NewValue)

	case "ENGINEER_RESOLUTION", "RESOLUTION":
		return ticketHistoryPDFFirst(h.Message, h.NewValue)

	case "CANCELLATION_REASON":
		return ticketHistoryPDFFirst(h.Message, h.NewValue)

	case "TICKET_OVERDUE",
		"STAFF_RESPONSE_OVERDUE",
		"STAFF_RESPONSE_LATE":
		return ticketHistoryPDFFirst(h.Message, h.NewValue)

	default:
		if message := ticketHistoryPDFValue(h.Message); message != "-" {
			return message
		}

		if h.OldValue != nil || h.NewValue != nil {
			return fmt.Sprintf(
				"%s: %s -> %s",
				h.FieldName,
				ticketHistoryPDFValue(h.OldValue),
				ticketHistoryPDFValue(h.NewValue),
			)
		}

		return ticketHistoryPDFFirst(h.NewValue)
	}
}
