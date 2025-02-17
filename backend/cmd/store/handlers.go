package main

import (
	"LuomuTori/internal/config"
	"LuomuTori/internal/log"
	"LuomuTori/internal/model"
	"LuomuTori/internal/model/view"
	"LuomuTori/internal/service/auth"
	"LuomuTori/internal/service/captcha"
	"LuomuTori/internal/service/dispute"
	"LuomuTori/internal/service/order"
	"LuomuTori/internal/service/payment"
	"LuomuTori/internal/service/pgp"
	"LuomuTori/internal/service/pledge"
	"LuomuTori/internal/service/product"
	"LuomuTori/internal/translate"
	"LuomuTori/internal/validate"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"

	moneropay "gitlab.com/moneropay/moneropay/v2/pkg/model"
)

const (
	maxMemory = 10 << 20
)

func (app *application) servePage(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		app.render(w, r, http.StatusOK, page, nil)
	}
}

func (app *application) index(w http.ResponseWriter, r *http.Request) {
	if user := app.loggedInUser(r); user != nil {
		http.Redirect(w, r, "/products", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func (app *application) captcha(w http.ResponseWriter, r *http.Request) {
	ca := captcha.New()
	app.sessionManager.Put(r.Context(), "captchaSolution", ca.Solution)

	w.Header().Set("Content-type", "image/png")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	if err := png.Encode(w, ca.Image); err != nil {
		app.serverError(w, err)
		return
	}
}

func (app *application) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Info.Println(err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(registerForm)
	if err := app.schemaDecoder.Decode(form, r.PostForm); err != nil {
		log.Info.Println(err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if !app.validateCaptcha(r.Context(), form.CaptchaAnswer) {
		form.SetError("Failed to solve captcha")
		app.renderInvalidForm(w, r, "register.html", form)
		return
	}

	form.CheckField(validate.AtleastNRunes(string(form.Password), 8), "Password", "Salasanassa tulee olla ainakin 8 merkkiä")
	form.CheckField(form.Password == form.PasswordCheck, "PasswordCheck", "Salasanat eivät täsmää")

	if !form.Valid() {
		app.renderInvalidForm(w, r, "register.html", form)
		return
	}

	_, err := auth.Register(app.db, form.Username, form.Password, form.PgpKey)
	if err != nil {
		if errors.Is(err, auth.ErrUsernameAlreadyRegistered) {
			form.SetError(fmt.Sprintf("user %s has been already registered", form.Username))
			app.renderInvalidForm(w, r, "register.html", form)
			return
		} else if errors.Is(err, auth.ErrInvalidPGPKey) {
			form.SetError("Invalid PGP key")
			app.renderInvalidForm(w, r, "register.html", form)
			return
		}
		app.serverError(w, err)
		return
	}

	app.addNotes(r.Context(), "Registered successfully!")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Info.Println(err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(loginForm)
	if err := app.schemaDecoder.Decode(form, r.PostForm); err != nil {
		log.Info.Println(err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if !app.validateCaptcha(r.Context(), form.CaptchaAnswer) {
		form.SetError("Failed to solve captcha")
		app.renderInvalidForm(w, r, "login.html", form)
		return
	}

	user, err := auth.Authenticate(app.db, form.Username, form.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			form.SetError("Invalid credentials")
			app.renderInvalidForm(w, r, "login.html", form)
			return
		} else if errors.Is(err, auth.ErrAccountIsBanned) {
			app.addErrorNotes(r.Context(), "Your account is banned")
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		app.serverError(w, err)
		return
	}

	if user.PgpKey == nil {
		ctx := r.Context()
		app.sessionManager.RenewToken(ctx)
		app.sessionManager.Put(ctx, "userID", user.ID)
		if _, err := model.M.User.UpdatePrevLogin(app.db, user.ID); err != nil {
			log.Error.Printf("failed to update previous login time: %s\n", err.Error())
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	token, err := auth.Generate2FAToken(20)
	if err != nil {
		app.serverError(w, err)
		return
	}

	ctx := r.Context()
	app.sessionManager.Put(ctx, "2fa_token", token)
	app.sessionManager.Put(ctx, "2fa_uid", user.ID)
	app.sessionManager.RenewToken(ctx)

	url := config.Addr + "/login/2fa?token=" + token
	message := fmt.Sprintf("---- LUOMUTORI 2-FA\n---- URL: %s\n---- HOW TO LOGIN?\n---- 1. Copy the url from above.\n---- 2. Browse to it in your browser.\n", url)
	encryptedMessage, err := pgp.EncryptMessage(*user.PgpKey, message)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"encryptedMessage": encryptedMessage,
	})

	app.render(w, r, http.StatusOK, "2fa.html", data)
}

func (app *application) handle2FA(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	loginToken := app.sessionManager.Pop(ctx, "2fa_token")

	if token == loginToken {
		uid, ok := app.sessionManager.Pop(ctx, "2fa_uid").(uuid.UUID)
		if !ok {
			app.serverError(w, fmt.Errorf("2FA Failed"))
			return
		}
		app.sessionManager.Put(ctx, "userID", uid)
		app.sessionManager.RenewToken(ctx)
		if _, err := model.M.User.UpdatePrevLogin(app.db, uid); err != nil {
			log.Error.Printf("failed to update previous login time: %s\n", err.Error())
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		app.clientError(w, http.StatusBadRequest)
	}
}

func (app *application) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	app.sessionManager.Pop(ctx, "userID")
	app.sessionManager.RenewToken(ctx)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (app *application) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Info.Println(err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(changePasswordForm)
	if err := app.schemaDecoder.Decode(form, r.PostForm); err != nil {
		log.Info.Println(err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validate.AtleastNRunes(string(form.NewPassword), 8), "NewPassword", "must contain atleast 8 characters")
	form.CheckField(form.NewPassword == form.NewPasswordCheck, "NewPasswordCheck", "passwords must match")

	user := app.loggedInUser(r)

	if !form.Valid() {
		wallet, err := model.M.Wallet.GetForUser(app.db, user.ID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := app.newTemplateData(r, map[string]any{"wallet": wallet})
		app.renderInvalidFormWithData(w, r, "settings.html", form, data)
		return
	}

	if _, err := auth.ChangePassword(app.db, user.Username, form.Password, form.NewPassword); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) || errors.Is(err, auth.ErrInvalidPassword) {
			form.SetError(err.Error())
			wallet, err := model.M.Wallet.GetForUser(app.db, user.ID)
			if err != nil {
				app.serverError(w, err)
				return
			}
			data := app.newTemplateData(r, map[string]any{"wallet": wallet})
			app.renderInvalidFormWithData(w, r, "settings.html", form, data)
			return
		}
		app.serverError(w, err)
		return
	}

	app.addNotes(r.Context(), "Password changed.")
	http.Redirect(w, r, "/user/settings", http.StatusSeeOther)
}

func (app *application) handleCreateListing(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(createListingForm)
	if err := app.schemaDecoder.Decode(form, r.PostForm); err != nil {
		log.Info.Printf("form decode failed %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validate.AtleastNRunes(form.Title, 3), "Title", "Invalid title")
	form.CheckField(validate.AtleastNRunes(form.Description, 3), "Description", "Invalid description")

	for _, pricing := range form.Pricings {
		bad := (pricing.Quantity == 0 && pricing.Price > 0) || pricing.Price < 0 || pricing.Quantity < 0
		form.CheckField(!bad, "Pricing", "Invalid pricing")
		if bad {
			break
		}
	}

	for _, deliveryMethod := range form.DeliveryMethods {
		bad := (deliveryMethod.Description == "" && deliveryMethod.Price > 0) || deliveryMethod.Price < 0
		form.CheckField(!bad, "deliveryMethod", "Invalid deliverymethod")
		if bad {
			break
		}
	}

	if !form.Valid() {
		data := app.newTemplateData(r, nil)
		data.Form = form
		app.render(w, r, http.StatusBadRequest, "create-listing.html", data)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		log.Info.Printf("unable to read image file %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dst, err := os.Create(filepath.Join("uploads/product-images", handler.Filename))
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := app.loggedInUser(r)

	product, err := product.Create(
		app.db,
		form.Title,
		form.Description,
		handler.Filename,
		form.Pricings,
		form.DeliveryMethods,
		user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.addNotes(r.Context(), "Listing created!")
	http.Redirect(w, r, fmt.Sprintf("/product?id=%s", product.ID), http.StatusSeeOther)
}

func (app *application) products(w http.ResponseWriter, r *http.Request) {
	products, err := view.V.Product.GetAll(app.db)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"products": products,
	})
	app.render(w, r, http.StatusOK, "products.html", data)
}

func (app *application) product(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Printf("Failed to parse id from url: %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	product, err := view.V.Product.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"product": product,
	})

	app.render(w, r, http.StatusOK, "product.html", data)
}

func (app *application) handleOrder(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(orderForm)
	if err := app.schemaDecoder.Decode(form, r.PostForm); err != nil {
		app.serverError(w, err)
		return
	}

	customer := app.loggedInUser(r)

	newOrder, err := order.Create(app.db, form.PriceID, form.DeliveryMethodID, customer.ID, form.Details)
	if err != nil {
		if errors.Is(err, order.ErrNotEnoughBalance) {
			app.addErrorNotes(r.Context(), "Not enough balance!")
			app.redirectBack(w, r)
			return
		} else if errors.Is(err, order.ErrCustomerIsVendor) {
			app.addErrorNotes(r.Context(), "You can't order your own product!")
			app.redirectBack(w, r)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	app.addNotes(r.Context(), "Order created!")
	http.Redirect(w, r, fmt.Sprintf("/order?id=%s", newOrder.ID), http.StatusSeeOther)
}

func (app *application) depositCallback(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.serverError(w, err)
		return
	}

	var data moneropay.CallbackResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		app.serverError(w, err)
		return
	}

	log.Info.Printf("deposit callback: Amount.total: %s, amount.unlocked: %s\n",
		payment.XMR2Decimal(data.Amount.Covered.Total), payment.XMR2Decimal(data.Amount.Covered.Unlocked))

	w.WriteHeader(http.StatusOK)
}

func (app *application) invoice(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Printf("Failed to parse id from url: %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	invoice, err := model.M.Invoice.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	info, err := view.V.Invoice.Get(app.db, invoice.ID)
	data := app.newTemplateData(r, map[string]any{"invoice": info})
	app.render(w, r, http.StatusOK, "invoice.html", data)
}

func (app *application) order(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Printf("Failed to parse id from url: %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)

	if !order.IsCustomer(app.db, user.ID, id) && !order.IsVendor(app.db, user.ID, id) {
		log.Info.Printf("User %s is not a customer or vendor of this order %s\n", user.ID, id)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	view, err := view.V.Order.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r, map[string]any{"order": view})
	app.render(w, r, http.StatusOK, "order.html", data)
}

func (app *application) ordersPlaced(w http.ResponseWriter, r *http.Request) {
	user := app.loggedInUser(r)

	infos, err := view.V.Order.GetAllForCustomer(app.db, user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r, map[string]any{"orders": infos})
	app.render(w, r, http.StatusOK, "orders-placed.html", data)
}

func (app *application) ordersIncoming(w http.ResponseWriter, r *http.Request) {
	user := app.loggedInUser(r)

	infos, err := view.V.Order.GetAllForVendor(app.db, user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data := app.newTemplateData(r, map[string]any{"orders": infos})
	app.render(w, r, http.StatusOK, "orders-incoming.html", data)
}

func (app *application) handleWithdrawal(w http.ResponseWriter, r *http.Request) {
	form := new(withdrawForm)
	err := app.decodeForm(r, form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validate.ValidXMRAddress(form.Address), "Address", "Not a valid XMR-address")
	form.CheckField(form.AmountFiat >= 10, "AmountFiat", "Minimum withdrawal amount is 10€")
	if !form.Valid() {
		app.addErrorNotes(r.Context(), "Minimum withdrawal amount is 10€!")
		app.renderInvalidForm(w, r, "wallet.html", form)
		return
	}

	user := app.loggedInUser(r)
	amount, err := payment.WithdrawFunds(app.db, user.ID, form.Address, payment.Fiat2XMR(form.AmountFiat))
	if errors.Is(err, payment.ErrNotEnoughBalanceToWithdraw) {
		form.SetError("Minimum withdrawal amount is 10€")
		app.addErrorNotes(r.Context(), "Not enough balance!")
		app.renderInvalidForm(w, r, "wallet.html", form)
		return
	} else if err != nil {
		app.serverError(w, err)
		return
	}

	log.Info.Printf("Withdrawal of %s XMR to %s initiated.\n", payment.XMR2Decimal(amount), form.Address)

	app.addNotes(r.Context(), "Withdrawal initiated successfully!")

	http.Redirect(w, r, "/user/wallet", http.StatusSeeOther)
}

func (app *application) handleRefund(w http.ResponseWriter, r *http.Request) {
	form := new(orderRefundForm)
	err := app.decodeForm(r, form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, form.OrderID) {
		log.Info.Printf("logged in user must be the vendor for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := order.Refund(app.db, form.OrderID); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/orders/incoming", http.StatusSeeOther)
}

func (app *application) review(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Println("unable to parse id")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsCustomer(app.db, user.ID, orderID) {
		log.Info.Printf("logged in user must be the customer for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	view, err := view.V.Order.Get(app.db, orderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{"order": view})
	app.render(w, r, http.StatusOK, "review.html", data)
}

func (app *application) handleReview(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(reviewForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Grade >= 0 && form.Grade <= 5, "grade", "grade must be in range [0, 5]")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		view, err := view.V.Order.Get(app.db, form.OrderID)
		if err != nil {
			app.serverError(w, err)
			return
		}
		data := app.newTemplateData(r, map[string]any{"order": view})

		app.renderInvalidFormWithData(w, r, fmt.Sprintf("/review?id=%s", form.OrderID), form, data)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsCustomer(app.db, user.ID, form.OrderID) {
		log.Info.Printf("logged in user must be the customer for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = order.Complete(app.db, form.OrderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	_, err = model.M.Review.Create(app.db, form.Grade, form.Message, form.OrderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.addNotes(r.Context(), "Review created. Thank you!")
	http.Redirect(w, r, "/orders/placed", http.StatusSeeOther)
}

func (app *application) dispute(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Println("unable to parse id")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsCustomer(app.db, user.ID, orderID) {
		log.Info.Printf("logged in user must be the customer for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	view, err := view.V.Order.Get(app.db, orderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{"order": view})
	app.render(w, r, http.StatusOK, "dispute.html", data)
}

func (app *application) handleDispute(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(disputeForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Claim != "", "claim", "claim can't be empty")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsCustomer(app.db, user.ID, form.OrderID) {
		log.Info.Printf("logged in user must be the customer for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := dispute.CreateDispute(app.db, form.OrderID, form.Claim); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/orders/placed", http.StatusSeeOther)
}

func (app *application) counterDispute(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Println("unable to parse id")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, orderID) {
		log.Info.Printf("logged in user must be the vendor for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	dispute, err := model.M.Dispute.GetForOrder(app.db, orderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	view, err := view.V.Dispute.Get(app.db, dispute.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"dispute": view,
		"order":   view.Order,
	})
	app.render(w, r, http.StatusOK, "counter-dispute.html", data)
}

func (app *application) handleCounterDispute(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(disputeForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Claim != "", "claim", "claim can't be empty")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, form.OrderID) {
		log.Info.Printf("logged in user must be the vendor for this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := dispute.CreateCounterDispute(app.db, form.OrderID, form.Claim); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/orders/incoming", http.StatusSeeOther)
}

func (app *application) handleVendorPledge(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("logo")
	if err != nil {
		log.Info.Printf("unable to read image file %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)

	// Require pgp key to become vendor
	if user.PgpKey == nil {
		app.addErrorNotes(r.Context(), "pgp is required from vendors!")
		app.redirectBack(w, r)
		return
	}

	logoFilename := user.Username + filepath.Ext(handler.Filename)

	dst, err := os.Create(filepath.Join("uploads/vendor-logos", logoFilename))
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if _, err := pledge.Create(app.db, user.ID, logoFilename); err != nil {
		if errors.Is(err, pledge.ErrNotEnoughBalance) {
			app.addErrorNotes(r.Context(), err.Error())
		} else if errors.Is(err, pledge.ErrUserIsAlreadyVendor) {
			app.addErrorNotes(r.Context(), err.Error())
		} else {
			app.serverError(w, err)
			return
		}
	} else {
		app.addNotes(r.Context(), "Congratulations for becoming vendor!")
	}

	app.redirectBack(w, r)
}

func (app *application) decline(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Println("unable to parse id")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, orderID) {
		log.Info.Printf("logged in user must be the vendor to decline this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	view, err := view.V.Order.Get(app.db, orderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{"order": view})
	app.render(w, r, http.StatusOK, "decline.html", data)
}

func (app *application) handleDecline(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(declineForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Reason != "", "reason", "reason can't be empty")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, form.OrderID) {
		log.Info.Printf("logged in user must be the vendor to decline this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := order.Decline(app.db, form.OrderID, form.Reason); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/orders/incoming", http.StatusSeeOther)
}

func (app *application) deliver(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Println("unable to parse id")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, orderID) {
		log.Info.Printf("logged in user must be the vendor to decline this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	view, err := view.V.Order.Get(app.db, orderID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{"order": view})
	app.render(w, r, http.StatusOK, "deliver.html", data)
}

func (app *application) handleDeliver(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(deliverForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Info != "", "Info", "info can't be empty")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	user := app.loggedInUser(r)
	if !order.IsVendor(app.db, user.ID, form.OrderID) {
		log.Info.Printf("logged in user must be the vendor to decline this order")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := order.Deliver(app.db, form.OrderID, form.Info); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/orders/incoming", http.StatusSeeOther)
}

func (app *application) handleProductDelete(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(productDeleteForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	product, err := model.M.Product.Get(app.db, form.ProductID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := app.loggedInUser(r)
	if user.ID != product.VendorID {
		log.Info.Printf("vendors can only delete their products")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := model.M.Product.Delete(app.db, product.ID); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, "/products", http.StatusSeeOther)
}

func (app *application) handleTicket(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(ticketForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Subject != "", "subject", "subject can't be empty")
	form.CheckField(form.Message != "", "message", "message can't be empty")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		app.renderInvalidForm(w, r, "create-ticket.html", form)
		return
	}

	user := app.loggedInUser(r)
	ticket, err := model.M.Ticket.Create(app.db, form.Subject, form.Message, user.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ticket/view?id=%s", ticket.ID), http.StatusSeeOther)
}

func (app *application) ticket(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Printf("unable to parse id from query url %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	ticket, err := view.V.Ticket.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := app.loggedInUser(r)
	if user.ID != ticket.Ticket.AuthorID {
		log.Info.Println("user must be the author of this ticket")
		app.clientError(w, http.StatusBadRequest)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"ticket": ticket,
	})
	app.render(w, r, http.StatusOK, "ticket.html", data)
}

func (app *application) tickets(w http.ResponseWriter, r *http.Request) {
	author := app.loggedInUser(r)
	tickets, err := model.M.Ticket.GetAllForAuthor(app.db, author.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"tickets": tickets,
	})
	app.render(w, r, http.StatusOK, "tickets.html", data)
}

func (app *application) handleTicketResponse(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Info.Printf("unable to parse form %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form := new(ticketResponseForm)

	err = app.schemaDecoder.Decode(form, r.PostForm)
	if err != nil {
		app.serverError(w, err)
		return
	}

	form.CheckField(form.Message != "", "message", "message can't be empty")

	if !form.Valid() {
		log.Info.Println("received invalid form")
		app.renderInvalidForm(w, r, "create-ticket.html", form)
		return
	}

	user := app.loggedInUser(r)
	ticket, err := model.M.Ticket.Get(app.db, form.TicketID)
	// TODO: Notify user what went wrong
	if err != nil || ticket.AuthorID != user.ID || !ticket.IsOpen {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := model.M.TicketResponse.Create(app.db, form.Message, form.TicketID, user.Username); err != nil {
		app.serverError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/ticket/view?id=%s", ticket.ID), http.StatusSeeOther)
}

func (app *application) toggleLang(w http.ResponseWriter, r *http.Request) {
	lang := app.sessionManager.GetString(r.Context(), "Lang")
	if lang == translate.En {
		app.sessionManager.Put(r.Context(), "Lang", "fi")
	} else {
		app.sessionManager.Put(r.Context(), "Lang", "en")
	}

	referer := r.Header.Get("referer")
	if referer != "" {
		http.Redirect(w, r, referer, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (app *application) vendor(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.URL.Query().Get("id"))
	if err != nil {
		log.Info.Printf("Failed to parse id from url: %s\n", err.Error())
		app.clientError(w, http.StatusBadRequest)
		return
	}

	vendor, err := view.V.Vendor.Get(app.db, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r, map[string]any{
		"vendor": vendor,
	})
	app.render(w, r, http.StatusOK, "vendor.html", data)
}
