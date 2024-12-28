package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type VendorPledge struct {
	ID           uuid.UUID
	Amount       uint64
	LogoFilename string
	UserID       uuid.UUID
	CreatedAt    time.Time
}

type VendorPledgeModel struct{}

func (m VendorPledgeModel) Create(ec db.ExecContext, amount uint64, logoFilename string, userID uuid.UUID) (*VendorPledge, error) {
	query := "INSERT INTO vendor_pledges (amount, logo_filename, user_id) VALUES($1, $2, $3) RETURNING id, created_at"

	pledge := &VendorPledge{
		Amount:       amount,
		LogoFilename: logoFilename,
		UserID:       userID,
	}

	err := ec.QueryRow(query, amount, logoFilename, userID).Scan(&pledge.ID, &pledge.CreatedAt)
	if err != nil {
		return nil, err
	}

	return pledge, nil
}

func (m VendorPledgeModel) GetForUser(ec db.ExecContext, userID uuid.UUID) (*VendorPledge, error) {
	query := "SELECT id, amount, logo_filename, created_at FROM vendor_pledges WHERE user_id=$1"

	pledge := &VendorPledge{
		UserID: userID,
	}

	err := ec.QueryRow(query, userID).Scan(&pledge.ID, &pledge.Amount, &pledge.LogoFilename, &pledge.CreatedAt)
	if err != nil {
		return nil, err
	}

	return pledge, nil

}
