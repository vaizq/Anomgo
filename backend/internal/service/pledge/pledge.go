package pledge

import (
	mydb "LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"LuomuTori/internal/service/payment"
	"database/sql"
	"github.com/google/uuid"

	"errors"
)

var (
	ErrNotEnoughBalance    = errors.New("Not enough balance")
	ErrUserIsAlreadyVendor = errors.New("User is already a vendor")
)

func Create(db *sql.DB, userID uuid.UUID, logoFilename string) (*model.VendorPledge, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	wallet, err := model.M.Wallet.GetForUser(tx, userID)
	if err != nil {
		return nil, err
	}

	pledgeAmount := payment.XMR

	// Check if this fails because of the balance >= $2 constraint
	if _, err := model.M.Wallet.ReduceBalance(tx, wallet.ID, pledgeAmount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotEnoughBalance
		}
		return nil, err
	}

	pledge, err := model.M.VendorPledge.Create(tx, pledgeAmount, logoFilename, userID)
	if err != nil {
		if mydb.ErrCode(err) == mydb.ErrCodeUniqueViolation {
			return nil, ErrUserIsAlreadyVendor
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return pledge, nil
}
