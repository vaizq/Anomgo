package order

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"LuomuTori/internal/service/payment"
	"database/sql"
	"errors"
	"github.com/google/uuid"
)

var (
	ErrNotEnoughBalance = errors.New("Not enough balance")
	ErrCustomerIsVendor = errors.New("Customer can't be vendor")
)

func Create(db *sql.DB, priceID, deliveryMethodID, customerID uuid.UUID, details string) (*model.Order, error) {
	price, err := model.M.Price.Get(db, priceID)
	if err != nil {
		return nil, err
	}

	delivery, err := model.M.DeliveryMethod.Get(db, deliveryMethodID)
	if err != nil {
		return nil, err
	}

	wallet, err := model.M.Wallet.GetForUser(db, customerID)
	if err != nil {
		return nil, err
	}

	xmrPrice := payment.Fiat2XMR(float64(price.Price + delivery.Price))
	if wallet.Balance >= xmrPrice {
		tx, err := db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		_, err = model.M.Wallet.ReduceBalance(tx, wallet.ID, xmrPrice)
		if err != nil {
			return nil, err
		}

		order, err := model.M.Order.Create(tx, priceID, deliveryMethodID, customerID, model.StatusPaid, details)
		if err != nil {
			return nil, err
		}

		if IsVendor(tx, customerID, order.ID) {
			return nil, ErrCustomerIsVendor
		}

		// Invoice is used to store the value in escrow
		if _, err := model.M.Invoice.Create(tx, wallet.Address, order.ID, payment.TakeCut(xmrPrice)); err != nil {
			return nil, err
		}

		if err := tx.Commit(); err != nil {
			return nil, err
		}

		return order, nil
	} else {
		return nil, ErrNotEnoughBalance
	}
}

func Complete(db *sql.DB, orderID uuid.UUID) error {
	invoice, err := model.M.Invoice.GetWithOrderID(db, orderID)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order, err := model.M.Order.UpdateStatus(tx, orderID, model.StatusCompleted)
	if err != nil {
		return err
	}

	vendor, err := model.M.Order.GetVendor(tx, order.ID)
	if err != nil {
		return err
	}

	wallet, err := model.M.Wallet.GetForUser(tx, vendor.ID)
	if err != nil {
		return err
	}

	if _, err = model.M.Wallet.AddBalance(tx, wallet.ID, invoice.XMRPrice); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func Refund(db *sql.DB, orderID uuid.UUID) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order, err := model.M.Order.UpdateStatus(tx, orderID, model.StatusDisputeSettled)
	if err != nil {
		return err
	}

	invoice, err := model.M.Invoice.GetForOrder(tx, order.ID)
	if err != nil {
		return err
	}

	customerWallet, err := model.M.Wallet.GetForUser(tx, order.CustomerID)
	if err != nil {
		return err
	}

	if _, err := model.M.Wallet.AddBalance(tx, customerWallet.ID, invoice.XMRPrice); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func Deliver(db *sql.DB, orderID uuid.UUID, info string) (*model.Order, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	order, err := model.M.Order.UpdateStatus(tx, orderID, model.StatusDelivered)
	if err != nil {
		return nil, err
	}

	if _, err := model.M.DeliveryInfo.Create(tx, info, orderID); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return order, nil

}

func Decline(db *sql.DB, orderID uuid.UUID, reason string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	order, err := model.M.Order.UpdateStatus(tx, orderID, model.StatusDeclined)
	if err != nil {
		return err
	}

	invoice, err := model.M.Invoice.GetForOrder(tx, order.ID)
	if err != nil {
		return err
	}

	customerWallet, err := model.M.Wallet.GetForUser(tx, order.CustomerID)
	if err != nil {
		return err
	}

	if _, err := model.M.Wallet.AddBalance(tx, customerWallet.ID, invoice.XMRPrice); err != nil {
		return err
	}

	if _, err := model.M.DeclineReason.Create(tx, reason, orderID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// Completes orders that have been delivered over 14 days ago,
// but have not been reviewed or disputed by the customer
func CompleteForgotten(db *sql.DB) error {
	query := `
		SELECT orders.id
		FROM orders 
		JOIN delivery_infos AS deliverys ON deliverys.order_id = orders.id
		WHERE orders.status = 'delivered' AND deliverys.created_at <= NOW() - INTERVAL '14 days'
	`

	rows, err := db.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id uuid.UUID
		err := rows.Scan(&id)
		if err != nil {
			return err
		}
		if err := Complete(db, id); err != nil {
			return err
		}
	}

	return nil
}

func IsCustomer(ec db.ExecContext, userID, orderID uuid.UUID) bool {
	order, err := model.M.Order.Get(ec, orderID)
	return err == nil && userID == order.CustomerID
}

func IsVendor(ec db.ExecContext, userID, orderID uuid.UUID) bool {
	vendor, err := model.M.Order.GetVendor(ec, orderID)
	return err == nil && userID == vendor.ID
}
