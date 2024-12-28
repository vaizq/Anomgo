package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

type Ticket struct {
	Ticket    *model.Ticket
	Responses []model.TicketResponse
}

type TicketView struct{}

func (tv TicketView) Get(ec db.ExecContext, id uuid.UUID) (*Ticket, error) {
	ticket, err := model.M.Ticket.Get(ec, id)
	if err != nil {
		return nil, err
	}

	responses, err := model.M.TicketResponse.GetAllForTicket(ec, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return &Ticket{
		Ticket:    ticket,
		Responses: responses,
	}, nil
}
