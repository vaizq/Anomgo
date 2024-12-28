package payment

import (
	"LuomuTori/internal/model"
	"database/sql"
	"github.com/google/uuid"
	moneropay "gitlab.com/moneropay/moneropay/v2/pkg/model"
)

func CreateInvoiceForDeposits(userID uuid.UUID) (*moneropay.ReceivePostResponse, error) {
	return moneropayReceivePost(0, "deposit", moneropayDepositCallbackURL(userID))
}

func HandleDeposits(db *sql.DB) error {
	wallets, err := model.M.Wallet.GetAll(db)
	if err != nil {
		return err
	}

	for _, w := range wallets {
		data, err := moneropayReceiveGet(w.Address)
		if err != nil {
			return err
		}

		if data.Amount.Covered.Unlocked > 0 && data.Amount.Covered.Unlocked == data.Amount.Covered.Total {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			if _, err := model.M.Wallet.AddBalance(db, w.ID, data.Amount.Covered.Unlocked); err != nil {
				return err
			}

			invoice, err := CreateInvoiceForDeposits(w.UserID)
			if err != nil {
				return err
			}

			if _, err = model.M.Wallet.UpdateAddress(db, w.ID, invoice.Address); err != nil {
				return err
			}

			if err := tx.Commit(); err != nil {
				return err
			}
		}
	}
	return nil
}
