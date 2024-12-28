package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
)

type Price struct {
	ID        uuid.UUID
	Quantity  int
	Price     int
	ProductID uuid.UUID
}

type PriceModel struct{}

func (pm PriceModel) Create(ec db.ExecContext, quantity int, price int, productID uuid.UUID) (*Price, error) {
	query := "INSERT INTO prices (quantity, price, product_id) VALUES ($1, $2, $3) RETURNING id"

	p := &Price{
		Quantity:  quantity,
		Price:     price,
		ProductID: productID,
	}

	err := ec.QueryRow(query, quantity, price, productID).Scan(&p.ID)
	if err != nil {
		return nil, err
	}

	return p, err
}

func (pm PriceModel) Get(ec db.ExecContext, id uuid.UUID) (*Price, error) {
	query := "SELECT quantity, price, product_id FROM prices WHERE id = $1"

	p := &Price{
		ID: id,
	}

	err := ec.QueryRow(query, id).Scan(&p.Quantity, &p.Price, &p.ProductID)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (pm PriceModel) GetAll(ec db.ExecContext, productID uuid.UUID) ([]Price, error) {
	query := "SELECT id, quantity, price FROM prices WHERE product_id = $1"

	rows, err := ec.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prices := make([]Price, 0)

	for rows.Next() {
		p := Price{
			ProductID: productID,
		}
		err = rows.Scan(&p.ID, &p.Quantity, &p.Price)
		if err != nil {
			return nil, err
		}

		prices = append(prices, p)
	}

	return prices, nil
}
