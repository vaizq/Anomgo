package view

import (
	"LuomuTori/internal/db"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
)

type Vendor struct {
	User         *model.User
	LogoFilename string
	Rating       Rating
	NumReviews   int
	NumDisputes  int
}

type VendorView struct{}

func (pv VendorView) Get(ec db.ExecContext, vendorID uuid.UUID) (*Vendor, error) {
	user, err := model.M.User.Get(ec, vendorID)
	if err != nil {
		return nil, err
	}

	// Just to make sure that vendorID is a vendor
	pledge, err := model.M.VendorPledge.GetForUser(ec, vendorID)
	if err != nil {
		return nil, err
	}

	reviews, err := V.Review.GetAllForVendor(ec, vendorID)
	if err != nil {
		return nil, err
	}

	orders, err := model.M.Order.GetAllForVendor(ec, vendorID)
	if err != nil {
		return nil, err
	}

	numDisputes := 0
	for _, order := range orders {
		if order.Status == model.StatusDisputed || order.Status == model.StatusDisputeCountered {
			numDisputes += 1
		}
	}

	return &Vendor{
		User:         user,
		LogoFilename: pledge.LogoFilename,
		Rating:       calculateRating(reviews),
		NumReviews:   len(reviews),
		NumDisputes:  numDisputes,
	}, nil
}
