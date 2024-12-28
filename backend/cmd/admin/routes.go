package main

import (
	"LuomuTori/internal/config"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"net/http"
)

func (app *application) route() http.Handler {
	r := httprouter.New()

	staticServer := http.FileServer(http.Dir(config.StaticDir))
	r.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static/", staticServer))

	cssServer := http.FileServer(http.Dir(config.CssDir))
	r.Handler(http.MethodGet, "/ui/css/*filepath", http.StripPrefix("/ui/css/", cssServer))

	r.HandlerFunc(http.MethodGet, "/", app.admin)
	r.HandlerFunc(http.MethodGet, "/dispute", app.dispute)
	r.HandlerFunc(http.MethodGet, "/ticket", app.ticket)

	r.HandlerFunc(http.MethodPost, "/delete", app.handleOperation)
	r.HandlerFunc(http.MethodPost, "/dispute", app.handleDispute)
	r.HandlerFunc(http.MethodPost, "/ticket", app.handleTicket)

	secure := alice.New(setSecureHeaders, app.logRequest, app.sessionManager.LoadAndSave)
	return secure.Then(r)
}
