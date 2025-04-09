package main

import (
	"LuomuTori/internal/config"
	"LuomuTori/internal/service/payment"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) routeInternal() http.Handler {
	r := httprouter.New()
	r.HandlerFunc(http.MethodPost, payment.DepositRoute, app.depositCallback)
	def := alice.New(app.logRequest)
	return def.Then(r)
}

func (app *application) route() http.Handler {
	r := httprouter.New()

	staticServer := http.FileServer(http.Dir(config.StaticDir))
	r.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", staticServer))

	cssServer := http.FileServer(http.Dir(config.CssDir))
	r.Handler(http.MethodGet, "/ui/css/*filepath", http.StripPrefix("/ui/css/", cssServer))

	uploadServer := http.FileServer(http.Dir(config.UploadDir))
	r.Handler(http.MethodGet, "/uploads/*filepath", http.StripPrefix("/uploads/", uploadServer))

	r.HandlerFunc(http.MethodGet, "/", app.index)
	r.HandlerFunc(http.MethodGet, "/register", app.servePage("register.html"))
	r.HandlerFunc(http.MethodGet, "/login", app.servePage("login.html"))
	r.HandlerFunc(http.MethodGet, "/support", app.servePage("support.html"))
	r.HandlerFunc(http.MethodGet, "/faq", app.servePage("faq.html"))
	r.HandlerFunc(http.MethodGet, "/login/2fa", app.handle2FA)
	r.HandlerFunc(http.MethodGet, "/captcha", app.captcha)

	r.HandlerFunc(http.MethodPost, "/register", app.handleRegister)
	r.HandlerFunc(http.MethodPost, "/login", app.handleLogin)
	r.HandlerFunc(http.MethodPost, "/lang/toggle", app.toggleLang)

	requireAuth := alice.New(app.requireAuth)
	requireVendor := alice.New(app.requireVendor)

	r.Handler(http.MethodGet, "/product", requireAuth.ThenFunc(app.product))
	r.Handler(http.MethodGet, "/products", requireAuth.ThenFunc(app.products))
	r.Handler(http.MethodGet, "/orders/invoice", requireAuth.ThenFunc(app.invoice))
	r.Handler(http.MethodGet, "/orders/placed", requireAuth.ThenFunc(app.ordersPlaced))
	r.Handler(http.MethodGet, "/orders/incoming", requireAuth.ThenFunc(app.ordersIncoming))
	r.Handler(http.MethodGet, "/orders/review", requireAuth.ThenFunc(app.review))
	r.Handler(http.MethodGet, "/orders/dispute", requireAuth.ThenFunc(app.dispute))
	r.Handler(http.MethodGet, "/order", requireAuth.ThenFunc(app.order))
	r.Handler(http.MethodGet, "/user/settings", requireAuth.ThenFunc(app.servePage("settings.html")))
	r.Handler(http.MethodGet, "/user/wallet", requireAuth.ThenFunc(app.servePage("wallet.html")))
	r.Handler(http.MethodGet, "/ticket/create", requireAuth.Then(app.servePage("create-ticket.html")))
	r.Handler(http.MethodGet, "/ticket/view/all", requireAuth.ThenFunc(app.tickets))
	r.Handler(http.MethodGet, "/ticket/view", requireAuth.ThenFunc(app.ticket))
	r.Handler(http.MethodGet, "/vendor", requireAuth.ThenFunc(app.vendor))

	r.Handler(http.MethodPost, "/logout", requireAuth.ThenFunc(app.handleLogout))
	r.Handler(http.MethodPost, "/orders/create", requireAuth.ThenFunc(app.handleOrder))
	r.Handler(http.MethodPost, "/orders/refund", requireAuth.ThenFunc(app.handleRefund))
	r.Handler(http.MethodPost, "/orders/review", requireAuth.ThenFunc(app.handleReview))
	r.Handler(http.MethodPost, "/orders/dispute", requireAuth.ThenFunc(app.handleDispute))
	r.Handler(http.MethodPost, "/user/withdrawal", requireAuth.ThenFunc(app.handleWithdrawal))
	r.Handler(http.MethodPost, "/vendor/pledge", requireAuth.ThenFunc(app.handleVendorPledge))
	r.Handler(http.MethodPost, "/user/change-password", requireAuth.ThenFunc(app.handleChangePassword))
	r.Handler(http.MethodPost, "/user/enable2fa", requireAuth.ThenFunc(app.handleEnable2FA))
	r.Handler(http.MethodPost, "/ticket/create", requireAuth.ThenFunc(app.handleTicket))
	r.Handler(http.MethodPost, "/ticket/response", requireAuth.ThenFunc(app.handleTicketResponse))

	r.Handler(http.MethodGet, "/vendor/create-listing", requireVendor.Then(app.servePage("create-listing.html")))
	r.Handler(http.MethodGet, "/orders/counter-dispute", requireVendor.ThenFunc(app.counterDispute))
	r.Handler(http.MethodGet, "/orders/deliver", requireVendor.ThenFunc(app.deliver))
	r.Handler(http.MethodGet, "/orders/decline", requireVendor.ThenFunc(app.decline))

	r.Handler(http.MethodPost, "/vendor/create-listing", requireVendor.ThenFunc(app.handleCreateListing))
	r.Handler(http.MethodPost, "/orders/counter-dispute", requireVendor.ThenFunc(app.handleCounterDispute))
	r.Handler(http.MethodPost, "/orders/deliver", requireVendor.ThenFunc(app.handleDeliver))
	r.Handler(http.MethodPost, "/orders/decline", requireVendor.ThenFunc(app.handleDecline))
	r.Handler(http.MethodPost, "/product/delete", requireVendor.ThenFunc(app.handleProductDelete))

	secure := alice.New(setSecureHeaders, app.logRequest, app.sessionManager.LoadAndSave)
	return secure.Then(r)
}
