package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
)

type Review struct {
	ID      uuid.UUID
	Grade   int
	Message string
	OrderID uuid.UUID
}

type ReviewModel struct{}

func (m ReviewModel) Create(ec db.ExecContext, grade int, message string, orderID uuid.UUID) (*Review, error) {
	query := "INSERT INTO reviews (grade, message, order_id) VALUES ($1, $2, $3) RETURNING id"

	r := &Review{
		Grade:   grade,
		Message: message,
		OrderID: orderID,
	}

	err := ec.QueryRow(query, grade, message, orderID).Scan(&r.ID)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (m ReviewModel) GetAuthor(ec db.ExecContext, reviewID uuid.UUID) (*User, error) {
	query := `
		SELECT users.id, users.username, users.created_at
		FROM users
		JOIN orders ON orders.customer_id = users.id 
		JOIN reviews ON reviews.order_id = orders.id
		WHERE reviews.id = $1;
		`

	u := &User{}
	err := ec.QueryRow(query, reviewID).Scan(&u.ID, &u.Username, &u.CreatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (m ReviewModel) GetAllForProduct(ec db.ExecContext, productID uuid.UUID) ([]Review, error) {
	query := `
		SELECT reviews.id, reviews.grade, reviews.message, reviews.order_id
		FROM reviews
		JOIN orders ON orders.id = reviews.order_id
		JOIN prices ON prices.id = orders.price_id
		WHERE prices.product_id = $1;
		`

	rows, err := ec.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)
	for rows.Next() {
		r := Review{}
		err := rows.Scan(&r.ID, &r.Grade, &r.Message, &r.OrderID)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (m ReviewModel) GetAllForVendor(ec db.ExecContext, vendorID uuid.UUID) ([]Review, error) {
	query := `
		SELECT reviews.id, reviews.grade, reviews.message, reviews.order_id
		FROM reviews
		JOIN orders ON reviews.order_id = orders.order_id
		JOIN prices ON orders.price_id = prices.id
		JOIN products ON prices.product_id = products.id
		WHERE products.vendor_id = $1;
		`

	rows, err := ec.Query(query, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)
	for rows.Next() {
		r := Review{}
		err := rows.Scan(&r.ID, &r.Grade, &r.Message, &r.OrderID)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, r)
	}

	return reviews, nil
}

func (m ReviewModel) Delete(ec db.ExecContext, id uuid.UUID) error {
	query := "DELETE FROM reviews WHERE id = $1"
	_, err := ec.Exec(query, id)
	return err
}
