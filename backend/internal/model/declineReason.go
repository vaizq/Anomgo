package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type DeclineReason struct {
	ID        uuid.UUID
	Reason    string
	OrderID   uuid.UUID
	CreatedAt time.Time
}

type DeclineReasonModel struct{}

func (m DeclineReasonModel) Create(ec db.ExecContext, reason string, orderID uuid.UUID) (*DeclineReason, error) {
	query := "INSERT INTO decline_reasons (reason, order_id) VALUES($1, $2) RETURNING id, created_at"

	dr := &DeclineReason{
		Reason:  reason,
		OrderID: orderID,
	}

	err := ec.QueryRow(query, reason, orderID).Scan(&dr.ID, &dr.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dr, err
}

func (m DeclineReasonModel) GetForOrder(ec db.ExecContext, orderID uuid.UUID) (*DeclineReason, error) {
	query := "SELECT id, reason, created_at FROM decline_reasons WHERE order_id=$1"

	dr := &DeclineReason{
		OrderID: orderID,
	}

	err := ec.QueryRow(query, orderID).Scan(&dr.ID, &dr.Reason, &dr.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dr, err
}
