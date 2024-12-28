package payment

import (
	"LuomuTori/internal/log"
	"LuomuTori/internal/model"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"gitlab.com/moneropay/go-monero/walletrpc"
	"time"
)

var (
	ErrNotEnoughBalanceToWithdraw = errors.New("Not enough balance to withdraw")
)

func WithdrawFunds(db *sql.DB, userID uuid.UUID, destinationAddress string, amount uint64) (uint64, error) {
	wallet, err := model.M.Wallet.GetForUser(db, userID)
	if err != nil {
		return 0, err
	}

	minAmount := Fiat2XMR(10)
	if amount < minAmount || wallet.Balance < amount {
		return 0, ErrNotEnoughBalanceToWithdraw
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	_, err = model.M.Wallet.ReduceBalance(tx, wallet.ID, amount)
	if err != nil {
		return 0, err
	}

	ourFee := Fiat2XMR(1)
	if _, err := model.M.Withdrawal.Create(tx, destinationAddress, amount-ourFee, model.WithdrawalPending); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return amount, nil
}

func HandleWithdrawals(db *sql.DB) error {
	if err := transferWithdrawals(db); err != nil {
		return err
	}

	if err := handleTransactions(db); err != nil {
		return err
	}

	return nil
}

func transferWithdrawals(db *sql.DB) error {
	ws, err := model.M.Withdrawal.GetAllWithStatus(db, model.WithdrawalPending)
	if err != nil {
		return err
	}
	if len(ws) == 0 {
		return nil
	}

	dsts := []walletrpc.Destination{}
	for _, w := range ws {
		dsts = append(dsts, walletrpc.Destination{
			Amount:  w.Amount,
			Address: w.DestAddress,
		})
	}

	// Set withdrawal statuses to processing
	{
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		for _, w := range ws {
			if _, err := model.M.Withdrawal.UpdateStatus(tx, w.ID, model.WithdrawalProcessing); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Make transaction
	data, err := moneropayTransferPost(dsts)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, w := range ws {
		if err := model.M.Withdrawal.Delete(tx, w.ID); err != nil {
			log.Error.Printf("Failed to delete withdrawal. ID: %s\n", w.ID)
			return err
		}
	}

	for _, hash := range data.TxHashList {
		if _, err := model.M.Transaction.Create(tx, hash); err != nil {
			log.Error.Printf("Failed to create transaction. TxHash: %s\n", hash)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Deletes confirmed transactions and redos failed transactions
func handleTransactions(db *sql.DB) error {
	threshold := time.Now().Add(-time.Minute * 10)
	txs, err := model.M.Transaction.GetAllAfter(db, threshold)
	if err != nil {
		return err
	}

	for _, tx := range txs {
		data, err := moneropayTransferGet(tx.Hash)
		if err != nil {
			return err
		}

		if data.Confirmations >= 10 {
			log.Info.Printf("tx %s succeeded\n", data.TxHash)
			if err := model.M.Transaction.Delete(db, tx.ID); err != nil {
				return err
			}
		} else if data.State == "failed" {
			log.Error.Printf("tx %s failed\n", data.TxHash)

			if err := model.M.Transaction.Delete(db, tx.ID); err != nil {
				return err
			}

			resp, err := moneropayTransferPost(data.Destinations)
			if err != nil {
				return err
			}

			dbtx, err := db.Begin()
			if err != nil {
				return err
			}
			defer dbtx.Rollback()

			for _, txHash := range resp.TxHashList {
				if _, err := model.M.Transaction.Create(dbtx, txHash); err != nil {
					return err
				}
			}

			if err := dbtx.Commit(); err != nil {
				return err
			}
		}
	}

	return nil
}
