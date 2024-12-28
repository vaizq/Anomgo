package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type DisputeOutcome string

const (
	OutcomeVendorWon   DisputeOutcome = "vendor won"
	OutcomeDraw        DisputeOutcome = "draw"
	OutcomeCustomerWon DisputeOutcome = "customer won"
)

type DisputeDecision struct {
	ID        uuid.UUID
	Outcome   DisputeOutcome
	Reason    string
	CreatedAt time.Time
}

type DisputeDecisionModel struct{}

func (m DisputeDecisionModel) Create(ec db.ExecContext, outcome DisputeOutcome, reason string, disputeID uuid.UUID) (*DisputeDecision, error) {
	query := "INSERT INTO dispute_decisions (outcome, reason, dispute_id) VALUES($1, $2, $3) RETURNING id, created_at"

	decision := &DisputeDecision{
		Outcome: outcome,
		Reason:  reason,
	}

	if err := ec.QueryRow(query, outcome, reason, disputeID).Scan(&decision.ID, &decision.CreatedAt); err != nil {
		return nil, err
	}

	return decision, nil
}
