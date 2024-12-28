package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type CounterDispute struct {
	ID        uuid.UUID
	Claim     string
	DisputeID uuid.UUID
	CreatedAt time.Time
}

type CounterDisputeModel struct{}

func (m CounterDisputeModel) Create(ec db.ExecContext, claim string, disputeID uuid.UUID) (*CounterDispute, error) {
	query := "INSERT INTO counter_disputes (claim, dispute_id) VALUES($1, $2) RETURNING id, created_at"

	counter := &CounterDispute{
		Claim:     claim,
		DisputeID: disputeID,
	}

	err := ec.QueryRow(query, claim, disputeID).Scan(&counter.ID, &counter.CreatedAt)
	if err != nil {
		return nil, err
	}
	return counter, err
}

func (m CounterDisputeModel) GetForDispute(ec db.ExecContext, disputeID uuid.UUID) (*CounterDispute, error) {
	query := "SELECT id, claim, created_at FROM counter_disputes WHERE dispute_id = $1"

	counter := &CounterDispute{
		DisputeID: disputeID,
	}

	err := ec.QueryRow(query, disputeID).Scan(&counter.ID, &counter.Claim, &counter.CreatedAt)
	if err != nil {
		return nil, err
	}
	return counter, err
}
