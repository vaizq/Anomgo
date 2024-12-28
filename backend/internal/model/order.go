package model

import (
	"LuomuTori/internal/db"
	"fmt"
	"github.com/google/uuid"
	"time"
)

type OrderStatus string

const (
	StatusPaid             OrderStatus = "paid"
	StatusDeclined         OrderStatus = "declined"
	StatusDelivered        OrderStatus = "delivered"
	StatusCompleted        OrderStatus = "completed"
	StatusDisputed         OrderStatus = "disputed"
	StatusDisputeCountered OrderStatus = "dispute countered"
	StatusDisputeSettled   OrderStatus = "dispute settled"
)

type Order struct {
	ID               uuid.UUID
	Status           OrderStatus
	Details          string
	PriceID          uuid.UUID
	DeliveryMethodID uuid.UUID
	CustomerID       uuid.UUID
	CreatedAt        time.Time
}

type OrderModel struct{}

func validStatus(_ OrderStatus) bool {
	return true
}

func (om OrderModel) Create(ec db.ExecContext, priceID uuid.UUID, deliveryMethodID uuid.UUID, customerID uuid.UUID, status OrderStatus, details string) (*Order, error) {
	if !validStatus(status) {
		return nil, fmt.Errorf("not a valid status: %s", status)
	}

	query := "INSERT INTO orders (status, details, price_id, delivery_method_id, customer_id) VALUES($1, $2, $3, $4, $5) RETURNING id, created_at"

	order := &Order{
		Status:           status,
		Details:          details,
		PriceID:          priceID,
		DeliveryMethodID: deliveryMethodID,
		CustomerID:       customerID,
	}

	err := ec.QueryRow(query, status, details, priceID, deliveryMethodID, customerID).Scan(&order.ID, &order.CreatedAt)
	if err != nil {
		return nil, err
	}
	return order, err
}

func (om OrderModel) Get(ec db.ExecContext, id uuid.UUID) (*Order, error) {
	query := "SELECT status, details, price_id, delivery_method_id, customer_id, created_at FROM orders WHERE id = $1"

	o := &Order{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&o.Status, &o.Details, &o.PriceID, &o.DeliveryMethodID, &o.CustomerID, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	return o, err
}

func (om OrderModel) GetAll(ec db.ExecContext, customerID uuid.UUID) ([]Order, error) {
	query := "SELECT id, status, details, price_id, delivery_method_id, created_at FROM orders WHERE customer_id = $1"

	rows, err := ec.Query(query, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)

	for rows.Next() {
		o := Order{
			CustomerID: customerID,
		}
		err := rows.Scan(&o.ID, &o.Status, &o.Details, &o.PriceID, &o.DeliveryMethodID, &o.CreatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (om OrderModel) GetAllWithStatus(ec db.ExecContext, status OrderStatus) ([]Order, error) {
	query := "SELECT id, details, price_id, delivery_method_id, created_at FROM orders WHERE status = $1"

	rows, err := ec.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)

	for rows.Next() {
		o := Order{
			Status: status,
		}
		err := rows.Scan(&o.ID, &o.Details, &o.PriceID, &o.DeliveryMethodID, &o.CreatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (om OrderModel) GetAllForVendor(ec db.ExecContext, vendorID uuid.UUID) ([]Order, error) {
	query := `
	SELECT orders.id, orders.status, orders.details, orders.price_id, orders.delivery_method_id, orders.customer_id, created_at 
	FROM orders
	JOIN prices ON orders.price_id = prices.id
	JOIN products ON prices.product_id = products.id
	WHERE products.vendor_id = $1
	`

	rows, err := ec.Query(query, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)

	for rows.Next() {
		o := Order{}
		err := rows.Scan(&o.ID, &o.Status, &o.Details, &o.DeliveryMethodID, &o.PriceID, &o.CustomerID, &o.CreatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, o)
	}

	return orders, nil
}

func (om OrderModel) UpdateStatus(ec db.ExecContext, id uuid.UUID, status OrderStatus) (*Order, error) {
	// TODO: There could be more sanity checks
	if !validStatus(status) {
		return nil, fmt.Errorf("not a valid status: %s", status)
	}

	query := func() string {
		switch status {
		case StatusCompleted:
			return fmt.Sprintf("UPDATE orders SET status = '%s' WHERE id = $1 AND status = 'delivered' RETURNING details, price_id, customer_id, created_at", status)
		case StatusDisputeSettled:
			return fmt.Sprintf("UPDATE orders SET status = '%s' WHERE id = $1 AND status = 'disputed' OR status = 'dispute countered' RETURNING details, price_id, customer_id, created_at", status)
		default:
			return fmt.Sprintf("UPDATE orders SET status = '%s' WHERE id = $1 RETURNING details, price_id, customer_id, created_at", status)
		}
	}()

	o := &Order{
		ID:     id,
		Status: status,
	}
	err := ec.QueryRow(query, id).Scan(&o.Details, &o.PriceID, &o.CustomerID, &o.CreatedAt)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (om OrderModel) GetVendor(ec db.ExecContext, orderID uuid.UUID) (*User, error) {
	query := `
		SELECT users.id, users.username, users.created_at, users.pgp_key
		FROM orders 
		JOIN prices ON prices.id = orders.price_id	
		JOIN products ON products.id = prices.product_id
		JOIN users ON users.id = products.vendor_id	
		WHERE orders.id = $1
	`

	u := &User{}
	err := ec.QueryRow(query, orderID).Scan(&u.ID, &u.Username, &u.CreatedAt, &u.PgpKey)
	if err != nil {
		return nil, err
	}

	return u, nil
}
