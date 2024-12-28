package dispute

import (
	"LuomuTori/internal/model"
	"database/sql"
	"github.com/google/uuid"
)

func CreateDispute(db *sql.DB, orderID uuid.UUID, claim string) (*model.Dispute, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := model.M.Order.UpdateStatus(tx, orderID, model.StatusDisputed); err != nil {
		return nil, err
	}

	dispute, err := model.M.Dispute.Create(tx, claim, orderID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return dispute, nil
}

func CreateCounterDispute(db *sql.DB, orderID uuid.UUID, claim string) (*model.CounterDispute, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	dispute, err := model.M.Dispute.GetForOrder(tx, orderID)
	if err != nil {
		return nil, err
	}

	if _, err := model.M.Order.UpdateStatus(tx, orderID, model.StatusDisputeCountered); err != nil {
		return nil, err
	}

	counterDispute, err := model.M.CounterDispute.Create(tx, claim, dispute.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return counterDispute, err
}

func CreateDisputeDecision(db *sql.DB, disputeID uuid.UUID, outcome model.DisputeOutcome, reason string) (*model.DisputeDecision, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	dispute, err := model.M.Dispute.Get(tx, disputeID)
	if err != nil {
		return nil, err
	}

	decision, err := model.M.DisputeDecision.Create(tx, outcome, reason, disputeID)
	if err != nil {
		return nil, err
	}

	order, err := model.M.Order.UpdateStatus(tx, dispute.OrderID, model.StatusDisputeSettled)
	if err != nil {
		return nil, err
	}

	invoice, err := model.M.Invoice.GetForOrder(tx, order.ID)
	if err != nil {
		return nil, err
	}

	vendor, err := model.M.Order.GetVendor(tx, order.ID)
	if err != nil {
		return nil, err
	}

	vendorWallet, err := model.M.Wallet.GetForUser(tx, vendor.ID)
	if err != nil {
		return nil, err
	}

	customerWallet, err := model.M.Wallet.GetForUser(tx, order.CustomerID)
	if err != nil {
		return nil, err
	}

	switch outcome {
	case model.OutcomeVendorWon:
		if _, err := model.M.Wallet.AddBalance(tx, vendorWallet.ID, invoice.XMRPrice); err != nil {
			return nil, err
		}
	case model.OutcomeDraw:
		split := invoice.XMRPrice / 2
		if _, err := model.M.Wallet.AddBalance(tx, vendorWallet.ID, split); err != nil {
			return nil, err
		}
		if _, err := model.M.Wallet.AddBalance(tx, customerWallet.ID, split); err != nil {
			return nil, err
		}
	case model.OutcomeCustomerWon:
		if _, err := model.M.Wallet.AddBalance(tx, customerWallet.ID, invoice.XMRPrice); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return decision, err
}
