package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type TicketResponse struct {
	ID         uuid.UUID
	Message    string
	TicketID   uuid.UUID
	AuthorName string
	CreatedAt  time.Time
}

type TicketResponseModel struct{}

func (m TicketResponseModel) Create(ec db.ExecContext, message string, ticketID uuid.UUID, authorName string) (*TicketResponse, error) {
	query := "INSERT INTO ticket_responses (message, ticket_id, author_name) VALUES($1, $2, $3) RETURNING id, created_at"

	t := &TicketResponse{
		Message:    message,
		TicketID:   ticketID,
		AuthorName: authorName,
	}

	if err := ec.QueryRow(query, message, ticketID, authorName).Scan(&t.ID, &t.CreatedAt); err != nil {
		return nil, err
	}

	return t, nil
}

func (m TicketResponseModel) GetAllForTicket(ec db.ExecContext, ticketID uuid.UUID) ([]TicketResponse, error) {
	query := "SELECT id, message, author_name, created_at FROM ticket_responses WHERE ticket_id=$1"

	rows, err := ec.Query(query, ticketID)
	if err != nil {
		return nil, err
	}

	ts := make([]TicketResponse, 0)

	for rows.Next() {
		t := TicketResponse{
			TicketID: ticketID,
		}

		if err := rows.Scan(&t.ID, &t.Message, &t.AuthorName, &t.CreatedAt); err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}

	return ts, nil

}
