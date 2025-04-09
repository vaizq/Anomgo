package main

import (
	"LuomuTori/internal/service/product"
	"LuomuTori/internal/validate"
	"github.com/google/uuid"
)

type CaptchaAnswer struct {
	X int `schema:"x"`
	Y int `schema:"y"`
}

type Captcha struct {
	CaptchaAnswer CaptchaAnswer
}

type registerForm struct {
	Username      string
	Password      string
	PasswordCheck string
	PgpKey        string
	Captcha
	validate.Validator
}

type loginForm struct {
	Username      string
	Password      string
	CaptchaAnswer CaptchaAnswer
	Captcha
	validate.Validator
}

type changePasswordForm struct {
	Password         string
	NewPassword      string
	NewPasswordCheck string
	validate.Validator
}

type enable2FAForm struct {
	PgpKey string
	validate.Validator
}

type createListingForm struct {
	Title           string
	Description     string
	DeliveryMethods []product.DeliveryMethod
	Pricings        []product.Pricing
	validate.Validator
}

type orderForm struct {
	PriceID          uuid.UUID
	DeliveryMethodID uuid.UUID
	Details          string
	validate.Validator
}

type orderRefundForm struct {
	OrderID uuid.UUID
	validate.Validator
}

type reviewForm struct {
	OrderID uuid.UUID
	Grade   int
	Message string
	validate.Validator
}

type disputeForm struct {
	OrderID uuid.UUID
	Claim   string
	validate.Validator
}

type withdrawForm struct {
	Address    string
	AmountFiat float64
	validate.Validator
}

type declineForm struct {
	OrderID uuid.UUID
	Reason  string
	validate.Validator
}

type deliverForm struct {
	OrderID uuid.UUID
	Info    string
	validate.Validator
}

type productDeleteForm struct {
	ProductID uuid.UUID
	validate.Validator
}

type ticketForm struct {
	Subject string
	Message string
	validate.Validator
}

type ticketResponseForm struct {
	TicketID uuid.UUID
	Message  string
	validate.Validator
}
