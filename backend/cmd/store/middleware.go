package main

import (
	"LuomuTori/internal/log"
	"LuomuTori/internal/model"
	"github.com/google/uuid"
	"net/http"
)

func setSecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/*
			w.Header().Set("Content-Security-Policy",
				"style-src 'self' fonts.googleapis.com unpkg.com; font-src fonts.gstatic.com; script-src 'self' unpkg.com; form-action 'self")
			w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "deny")
			w.Header().Set("X-XSS-Protection", "0")
		*/
		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.sessionManager.Exists(r.Context(), "userID") {
			next.ServeHTTP(w, r)
		} else {
			app.clientError(w, http.StatusUnauthorized)
		}
	})
}

func (app *application) requireVendor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if uid, ok := app.sessionManager.Get(r.Context(), "userID").(uuid.UUID); ok && model.M.User.IsVendor(app.db, uid) {
			next.ServeHTTP(w, r)
		} else {
			app.clientError(w, http.StatusUnauthorized)
		}
	})
}
