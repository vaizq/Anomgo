package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
)

type Product struct {
	Product         *model.Product
	Prices          []model.Price
	DeliveryMethods []model.DeliveryMethod
	Vendor          *Vendor
	Reviews         []Review
	Rating          Rating
	NumReviews      int
}

type ProductView struct{}

func (pv ProductView) Get(ec db.ExecContext, productID uuid.UUID) (*Product, error) {
	product, err := model.M.Product.Get(ec, productID)
	if err != nil {
		return nil, err
	}

	prices, err := model.M.Price.GetAll(ec, productID)
	if err != nil {
		return nil, err
	}

	dms, err := model.M.DeliveryMethod.GetAllForProduct(ec, productID)
	if err != nil {
		return nil, err
	}

	vendor, err := V.Vendor.Get(ec, product.VendorID)
	if err != nil {
		return nil, err
	}

	reviews, err := V.Review.GetAllForProduct(ec, productID)
	if err != nil {
		return nil, err
	}

	return &Product{
		Product:         product,
		Prices:          prices,
		DeliveryMethods: dms,
		Vendor:          vendor,
		Reviews:         reviews,
		Rating:          calculateRating(reviews),
		NumReviews:      len(reviews),
	}, nil
}

func (pv ProductView) GetAll(ec db.ExecContext) ([]Product, error) {
	products, err := model.M.Product.GetAll(ec)
	if err != nil {
		return nil, err
	}

	res := make([]Product, 0, len(products))

	for _, product := range products {
		p, err := pv.Get(ec, product.ID)
		if err != nil {
			return nil, err
		}
		res = append(res, *p)
	}
	return res, nil
}
