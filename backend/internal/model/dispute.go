package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type Dispute struct {
	ID        uuid.UUID
	Claim     string
	OrderID   uuid.UUID
	CreatedAt time.Time
}

type DisputeModel struct{}

func (m DisputeModel) Create(ec db.ExecContext, claim string, orderID uuid.UUID) (*Dispute, error) {
	query := "INSERT INTO disputes (claim, order_id) VALUES($1, $2) RETURNING id, created_at"

	dispute := &Dispute{
		Claim:   claim,
		OrderID: orderID,
	}

	err := ec.QueryRow(query, claim, orderID).Scan(&dispute.ID, &dispute.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dispute, err
}

func (m DisputeModel) Get(ec db.ExecContext, id uuid.UUID) (*Dispute, error) {
	query := "SELECT claim, order_id, created_at FROM disputes WHERE id = $1"

	dispute := &Dispute{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&dispute.Claim, &dispute.OrderID, &dispute.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dispute, err
}

func (m DisputeModel) GetAll(ec db.ExecContext) ([]Dispute, error) {
	query := "SELECT id, claim, order_id, created_at FROM disputes"

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}

	disputes := []Dispute{}

	for rows.Next() {
		d := Dispute{}
		if err := rows.Scan(&d.ID, &d.Claim, &d.OrderID, &d.CreatedAt); err != nil {
			return nil, err
		}
		disputes = append(disputes, d)
	}

	return disputes, nil
}

func (m DisputeModel) GetForOrder(ec db.ExecContext, orderID uuid.UUID) (*Dispute, error) {
	query := "SELECT id, claim, created_at FROM disputes WHERE order_id = $1"

	dispute := &Dispute{
		OrderID: orderID,
	}

	err := ec.QueryRow(query, orderID).Scan(&dispute.ID, &dispute.Claim, &dispute.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dispute, err
}
