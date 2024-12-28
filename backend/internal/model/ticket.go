package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type Ticket struct {
	ID        uuid.UUID
	Subject   string
	Message   string
	AuthorID  uuid.UUID
	IsOpen    bool
	CreatedAt time.Time
}

type TicketModel struct{}

func (m TicketModel) Create(ec db.ExecContext, subject string, message string, authorID uuid.UUID) (*Ticket, error) {
	query := "INSERT INTO tickets (subject, message, author_id) VALUES($1, $2, $3) RETURNING id, is_open, created_at"

	t := &Ticket{
		Subject:  subject,
		Message:  message,
		AuthorID: authorID,
	}

	if err := ec.QueryRow(query, subject, message, authorID).Scan(&t.ID, &t.IsOpen, &t.CreatedAt); err != nil {
		return nil, err
	}

	return t, nil
}

func (m TicketModel) Get(ec db.ExecContext, id uuid.UUID) (*Ticket, error) {
	query := "SELECT subject, message, author_id, is_open, created_at FROM tickets WHERE id=$1"

	t := &Ticket{
		ID: id,
	}

	if err := ec.QueryRow(query, id).Scan(&t.Subject, &t.Message, &t.AuthorID, &t.IsOpen, &t.CreatedAt); err != nil {
		return nil, err
	}

	return t, nil
}

func (m TicketModel) GetAll(ec db.ExecContext) ([]Ticket, error) {
	query := "SELECT id, subject, message, author_id, is_open, created_at FROM tickets"

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}

	tickets := []Ticket{}

	for rows.Next() {
		t := Ticket{}
		if err := rows.Scan(&t.ID, &t.Subject, &t.Message, &t.AuthorID, &t.IsOpen, &t.CreatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}

	return tickets, nil
}

func (m TicketModel) GetAllForAuthor(ec db.ExecContext, authorID uuid.UUID) ([]Ticket, error) {
	query := "SELECT id, subject, message, is_open, created_at FROM tickets WHERE author_id=$1"

	rows, err := ec.Query(query, authorID)
	if err != nil {
		return nil, err
	}

	ts := make([]Ticket, 0)

	for rows.Next() {
		t := Ticket{
			AuthorID: authorID,
		}

		if err := rows.Scan(&t.ID, &t.Subject, &t.Message, &t.IsOpen, &t.CreatedAt); err != nil {
			return nil, err
		}

		ts = append(ts, t)
	}

	return ts, nil
}

func (m TicketModel) Close(ec db.ExecContext, id uuid.UUID) (*Ticket, error) {
	query := "UPDATE tickets SET is_open = FALSE WHERE id = $1 RETURNING subject, message, author_id, created_at"

	t := &Ticket{
		ID:     id,
		IsOpen: false,
	}

	if err := ec.QueryRow(query, id).Scan(&t.Subject, &t.Message, &t.AuthorID, &t.CreatedAt); err != nil {
		return nil, err
	}

	return t, nil
}
