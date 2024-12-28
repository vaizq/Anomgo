package model

import (
	"LuomuTori/internal/db"
	"github.com/google/uuid"
)

type DeliveryMethod struct {
	ID          uuid.UUID
	Description string
	Price       int
	ProductID   uuid.UUID
}

type DeliveryMethodModel struct{}

func (pm DeliveryMethodModel) Create(ec db.ExecContext, description string, price int, productID uuid.UUID) (*DeliveryMethod, error) {
	query := "INSERT INTO delivery_methods (description, price, product_id) VALUES ($1, $2, $3) RETURNING id"

	dm := &DeliveryMethod{
		Description: description,
		Price:       price,
		ProductID:   productID,
	}

	if err := ec.QueryRow(query, description, price, productID).Scan(&dm.ID); err != nil {
		return nil, err
	}

	return dm, nil
}

func (pm DeliveryMethodModel) Get(ec db.ExecContext, id uuid.UUID) (*DeliveryMethod, error) {
	query := "SELECT description, price, product_id FROM delivery_methods WHERE id=$1"

	dm := &DeliveryMethod{
		ID: id,
	}

	if err := ec.QueryRow(query, id).Scan(&dm.Description, &dm.Price, &dm.ProductID); err != nil {
		return nil, err
	}

	return dm, nil
}

func (pm DeliveryMethodModel) GetForOrder(ec db.ExecContext, orderID uuid.UUID) (*DeliveryMethod, error) {
	query := `
		SELECT dm.id, dm.description, dm.price, dm.product_id 
		FROM delivery_methods AS dm
		JOIN orders ON orders.delivery_method_id = dm.id
		WHERE orders.id=$1
	`

	dm := &DeliveryMethod{}

	if err := ec.QueryRow(query, orderID).Scan(&dm.ID, &dm.Description, &dm.Price, &dm.ProductID); err != nil {
		return nil, err
	}

	return dm, nil
}

func (pm DeliveryMethodModel) GetAllForProduct(ec db.ExecContext, productID uuid.UUID) ([]DeliveryMethod, error) {
	query := "SELECT id, description, price FROM delivery_methods WHERE product_id=$1"

	rows, err := ec.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dms := make([]DeliveryMethod, 0)
	for rows.Next() {
		dm := DeliveryMethod{
			ProductID: productID,
		}

		if err := rows.Scan(&dm.ID, &dm.Description, &dm.Price); err != nil {
			return nil, err
		}

		dms = append(dms, dm)
	}

	return dms, err
}
