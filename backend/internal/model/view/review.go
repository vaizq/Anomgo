package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
)

type Review struct {
	Review *model.Review
	Author *model.User
}

type ReviewView struct{}

func (rv ReviewView) GetAllForProduct(ec db.ExecContext, productID uuid.UUID) ([]Review, error) {
	query := `
		SELECT reviews.id, reviews.grade, reviews.message, reviews.order_id, 
			authors.id, authors.username, authors.created_at
		FROM reviews 
		JOIN orders ON orders.id = reviews.order_id
		JOIN users AS authors ON authors.id = orders.customer_id
		JOIN prices ON prices.id = orders.price_id 
		JOIN products ON products.id = prices.product_id
		WHERE products.id = $1;
		`

	rows, err := ec.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)
	for rows.Next() {
		r := &model.Review{}
		a := &model.User{}
		err = rows.Scan(&r.ID, &r.Grade, &r.Message, &r.OrderID, &a.ID, &a.Username, &a.CreatedAt)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, Review{
			Review: r,
			Author: a,
		})
	}

	return reviews, nil
}

func (rv ReviewView) GetAllForVendor(ec db.ExecContext, vendorID uuid.UUID) ([]Review, error) {
	query := `
		SELECT reviews.id, reviews.grade, reviews.message, reviews.order_id, 
			authors.id, authors.username, authors.created_at
		FROM products
		JOIN prices ON prices.product_id = products.id
		JOIN orders ON orders.price_id = prices.id
		JOIN reviews ON reviews.order_id = orders.id
		JOIN users AS authors ON authors.id = orders.customer_id
		WHERE products.vendor_id = $1
		`

	rows, err := ec.Query(query, vendorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := make([]Review, 0)
	for rows.Next() {
		r := &model.Review{}
		a := &model.User{}
		err = rows.Scan(&r.ID, &r.Grade, &r.Message, &r.OrderID, &a.ID, &a.Username, &a.CreatedAt)
		if err != nil {
			return nil, err
		}

		reviews = append(reviews, Review{
			Review: r,
			Author: a,
		})
	}

	return reviews, nil
}
