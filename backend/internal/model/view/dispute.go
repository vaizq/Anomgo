package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
)

type Dispute struct {
	Dispute *model.Dispute
	Order   *Order
}

type DisputeView struct{}

func (v DisputeView) Get(ec db.ExecContext, id uuid.UUID) (*Dispute, error) {
	dispute, err := model.M.Dispute.Get(ec, id)
	if err != nil {
		return nil, err
	}

	orderView, err := V.Order.Get(ec, dispute.OrderID)
	if err != nil {
		return nil, err
	}

	return &Dispute{
		Dispute: dispute,
		Order:   orderView,
	}, nil
}
