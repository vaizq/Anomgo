package product

import (
	"LuomuTori/internal/model"
	"database/sql"
	"github.com/google/uuid"
)

type Pricing struct {
	Quantity int
	Price    int
}

type DeliveryMethod struct {
	Description string
	Price       int
}

func Create(
	db *sql.DB,
	title string,
	description string,
	imageFile string,
	pricings []Pricing,
	deliveryMethods []DeliveryMethod,
	vendorID uuid.UUID) (*model.Product, error) {

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	product, err := model.M.Product.Create(tx, title, description, imageFile, vendorID)
	if err != nil {
		return nil, err
	}

	for _, pricing := range pricings {
		if pricing.Quantity <= 0 {
			continue
		}
		if _, err := model.M.Price.Create(tx, pricing.Quantity, pricing.Price, product.ID); err != nil {
			return nil, err
		}
	}

	for _, dm := range deliveryMethods {
		if len(dm.Description) == 0 {
			continue
		}
		if _, err := model.M.DeliveryMethod.Create(tx, dm.Description, dm.Price, product.ID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return product, nil
}
