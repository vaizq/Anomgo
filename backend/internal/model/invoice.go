package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
	"time"
)

type Invoice struct {
	ID        uuid.UUID
	Address   string
	OrderID   uuid.UUID
	XMRPrice  uint64
	CreatedAt time.Time
}

type InvoiceModel struct{}

func (m *InvoiceModel) Create(ec db.ExecContext, address string, orderID uuid.UUID, xmrPrice uint64) (*Invoice, error) {
	query := "INSERT INTO invoices (address, order_id, xmr_price) VALUES($1, $2, $3) RETURNING id, created_at"

	invoice := &Invoice{
		Address:  address,
		OrderID:  orderID,
		XMRPrice: xmrPrice,
	}

	err := ec.QueryRow(query, address, orderID, xmrPrice).Scan(&invoice.ID, &invoice.CreatedAt)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (m InvoiceModel) Get(ec db.ExecContext, id uuid.UUID) (*Invoice, error) {
	query := "SELECT address, order_id, xmr_price, created_at FROM invoices WHERE id = $1"
	invoice := &Invoice{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&invoice.Address, &invoice.OrderID, &invoice.XMRPrice, &invoice.CreatedAt)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (m InvoiceModel) GetForOrder(ec db.ExecContext, orderID uuid.UUID) (*Invoice, error) {
	query := "SELECT id, address, xmr_price, created_at FROM invoices WHERE order_id = $1"
	invoice := &Invoice{
		OrderID: orderID,
	}

	err := ec.QueryRow(query, orderID).Scan(&invoice.ID, &invoice.Address, &invoice.XMRPrice, &invoice.CreatedAt)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}

func (m InvoiceModel) GetAll(ec db.ExecContext) ([]Invoice, error) {
	query := "SELECT id, address, order_id, xmr_price, created_at FROM invoices"

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]Invoice, 0)

	for rows.Next() {
		invoice := Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.Address, &invoice.OrderID, &invoice.XMRPrice, &invoice.CreatedAt); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (m InvoiceModel) GetAllUnpaid(ec db.ExecContext) ([]Invoice, error) {
	query := `
		SELECT invoices.id, invoices.address, invoices.order_id, invoices.xmr_price, invoices.created_at 
		FROM invoices
		JOIN orders ON orders.id = invoices.order_id
		WHERE orders.status = 'pending'
	`

	rows, err := ec.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]Invoice, 0)

	for rows.Next() {
		invoice := Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.Address, &invoice.OrderID, &invoice.XMRPrice, &invoice.CreatedAt); err != nil {
			return nil, err
		}

		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func (m InvoiceModel) Delete(ec db.ExecContext, id uuid.UUID) error {
	query := "DELETE FROM invoices WHERE id = $1"
	_, err := ec.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func (m InvoiceModel) GetWithOrderID(ec db.ExecContext, orderID uuid.UUID) (*Invoice, error) {
	query := "SELECT id, address, xmr_price, created_at FROM invoices WHERE order_id = $1"
	invoice := &Invoice{
		OrderID: orderID,
	}

	err := ec.QueryRow(query, orderID).Scan(&invoice.ID, &invoice.Address, &invoice.XMRPrice, &invoice.CreatedAt)
	if err != nil {
		return nil, err
	}
	return invoice, nil
}
