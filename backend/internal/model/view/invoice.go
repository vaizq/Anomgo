package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
)

type Invoice struct {
	Invoice *model.Invoice
	Order   *Order
}

type InvoiceView struct{}

func (iv InvoiceView) Get(ec db.ExecContext, id uuid.UUID) (*Invoice, error) {
	invoice, err := model.M.Invoice.Get(ec, id)
	if err != nil {
		return nil, err
	}

	order, err := V.Order.Get(ec, invoice.OrderID)
	if err != nil {
		return nil, err
	}

	return &Invoice{
		Invoice: invoice,
		Order:   order,
	}, nil
}
